// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

// This benchmark is plan to profile blockchain related processing and
// searching. The benchmark base on RegTest and chain height should higher than
// 253582, the all data should be placed on this directory with name
// "elastos_test". ProcessBlock sub test will load 253583 block and confirm raw
// data, the raw data should be named "block.dat" and "confirm.dat",
// and place in the directory "elastos_test".
// To run this benchmark only, please type the flowing command on you command
// line: go test -run=nonthingplease -benchonly="true" -bench=.

package blockchain

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/common/config"
	"github.com/elastos/Elastos.ELA/common/log"
	"github.com/elastos/Elastos.ELA/core/types"
	"github.com/elastos/Elastos.ELA/core/types/payload"
	"github.com/elastos/Elastos.ELA/database"
	"github.com/elastos/Elastos.ELA/dpos/state"
	"github.com/elastos/Elastos.ELA/elanet/pact"
	"github.com/elastos/Elastos.ELA/utils/test"
)

const (
	incomingBlock   = "block.dat"
	incomingConfirm = "confirm.dat"

	chainHeight = uint32(253582)
)

var (
	params        = config.DefaultParams.RegNet()
	originLedger  *Ledger
	originAddress common.Uint168

	flagParsed    bool
	isBenchOnly   bool
	benchOnlyFlag = flag.String("benchonly", "",
		"used for workbench only initialize")

	stopHash              = initStopHash()
	randomBlockHashes     = initRandomBlockHashes()
	continuousBlockHashes = initContinuousBlockHashes()
	chain                 = benchBegin()
	newChain              = newBlockChain()
)

func BenchmarkBlockChain_PersistBlocks(b *testing.B) {
	chainStore := newChain.db.(*ChainStore)
	for i := chainHeight - 30000; i < chainHeight; i++ {
		blockHash, _ := DefaultLedger.Store.GetBlockHash(i)
		block, _ := DefaultLedger.Store.GetBlock(blockHash)

		chainStore.NewBatch()
		if err := chainStore.persistTrimmedBlock(block); err != nil {
			b.Error(err)
		}
		if err := chainStore.persistBlockHash(block); err != nil {
			b.Error(err)
		}
		if err := chainStore.persistCurrentBlock(block); err != nil {
			b.Error(err)
		}
		if err := chainStore.BatchCommit(); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBlockChain_FFLDBPersistBlocks(b *testing.B) {
	newChain.Nodes = make([]*BlockNode, chainHeight-30000)
	for i := chainHeight - 30000; i < chainHeight; i++ {
		blockHash, _ := DefaultLedger.Store.GetBlockHash(i)
		block, _ := DefaultLedger.Store.GetBlock(blockHash)

		newNode := NewBlockNode(&block.Header, &blockHash)
		if err := newChain.db.GetFFLDB().SaveBlock(block, newNode,
			nil, time.Unix(int64(block.Timestamp), 0)); err != nil {
			b.Error(err)
		}
		newChain.Nodes = append(newChain.Nodes, newNode)
	}
}

func BenchmarkBlockChain_GetBlocks(b *testing.B) {
	for i := chainHeight - 30000; i < chainHeight; i++ {
		blockHash, _ := DefaultLedger.Store.GetBlockHash(i)
		_, err := DefaultLedger.Store.GetBlock(blockHash)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBlockChain_FFLDBGetBlocks(b *testing.B) {
	for i := chainHeight - 30000; i < chainHeight; i++ {
		blockHash := DefaultLedger.Blockchain.Nodes[i].Hash
		_, err := DefaultLedger.Store.GetFFLDB().GetBlock(*blockHash)
		if err != nil {
			b.Error(err)
		}
	}
	// Sleep one second to ensure that only one test is performed.
	time.Sleep(time.Second)
}

func BenchmarkBlockChain_GetTransactions(b *testing.B) {
	for i := chainHeight - 30000; i < chainHeight; i++ {
		blockHash, _ := DefaultLedger.Store.GetBlockHash(i)
		block, err := DefaultLedger.Store.GetBlock(blockHash)
		if err != nil {
			b.Error(err)
		}
		for _, tx := range block.Transactions {
			_, _, err := DefaultLedger.Store.GetTransaction(tx.Hash())
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func BenchmarkBlockChain_CheckBlockContext(b *testing.B) {
	block, _ := newBlock()
	for i := 0; i < 10; i++ {
		if err := chain.CheckBlockContext(block, chain.BestChain); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBlockChain_ProcessBlock(b *testing.B) {
	_, _, err := chain.ProcessBlock(newBlock())
	if err != nil {
		b.Error(err)
	}
	// Sleep one second to ensure that only one test is performed.
	time.Sleep(time.Second)
}

func BenchmarkBlockChain_HaveBlock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range randomBlockHashes {
			chain.HaveBlock(v)
		}
	}
}

func BenchmarkBlockChain_GetDposBlockByHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range randomBlockHashes {
			chain.GetDposBlockByHash(*v)
		}
	}
}

func BenchmarkBlockChain_LocateBlocks(b *testing.B) {
	for i := 0; i < b.N; i++ {
		locator := make([]*common.Uint256, 0)
		for _, v := range continuousBlockHashes {
			locator = append(locator, v)
		}
		chain.LocateBlocks(locator, stopHash, pact.MaxBlocksPerMsg)
	}
}

func BenchmarkBlockChain_End(b *testing.B) {
	if chain != nil {
		chain.db.Close()
		chain = nil
	}

	if newChain != nil {
		newChain.db.Close()
		newChain = nil
	}

	DefaultLedger = originLedger
	FoundationAddress = originAddress
}

func newChainStore(dataDir string, dbDir string, genesisBlock *types.Block) (IChainStore, error) {
	db, err := NewLevelDB(filepath.Join(dataDir, dbDir, "chain"))
	if err != nil {
		return nil, err
	}

	fdb, err := NewChainStoreFFLDB(filepath.Join(dataDir, dbDir))
	if err != nil {
		return nil, err
	}

	s := &ChainStore{
		IStore:           db,
		fflDB:            fdb,
		blockHashesCache: make([]common.Uint256, 0, BlocksCacheSize),
		blocksCache:      make(map[common.Uint256]*types.Block),
	}

	s.init(genesisBlock)

	return s, nil
}

func newBlockChain() *BlockChain {
	log.NewDefault(test.NodeLogPath, 0, 0, 0)
	chainStore, err := newChainStore(test.DataPath, "ffldb", params.GenesisBlock)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	chain, err := New(chainStore, params, nil, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return chain
}

func benchBegin() *BlockChain {
	if !hasWorkbenchOnlyFlag() {
		return nil
	}

	log.NewDefault(test.NodeLogPath, 0, 0, 0)
	log.Info("Roll back to: ", chainHeight)
	rollbackTo(chainHeight, params)

	log.Info("New chain store.")
	chainStore, err := NewChainStore(test.DataPath, params.GenesisBlock)
	if err != nil {
		fmt.Println("create new chain store failed, ", err)
		return nil
	}

	originAddress = FoundationAddress
	FoundationAddress = params.Foundation

	log.Info("New arbitrator.")
	arbiters, _ := state.NewArbitrators(params, nil,
		func(programHash common.Uint168) (common.Fixed64,
			error) {
			amount := common.Fixed64(0)
			utxos, err := DefaultLedger.Store.
				GetUnspentFromProgramHash(programHash, config.ELAAssetID)
			if err != nil {
				return amount, err
			}
			for _, utxo := range utxos {
				amount += utxo.Value
			}
			return amount, nil
		})
	arbiters.RegisterFunction(chainStore.GetHeight, func(height uint32) (*types.Block, error) {
		hash, err := chainStore.GetBlockHash(height)
		if err != nil {
			return nil, err
		}
		block, err := chainStore.GetBlock(hash)
		if err != nil {
			return nil, err
		}
		CalculateTxsFee(block)
		return block, nil
	}, DefaultLedger.Blockchain.UTXOCache.GetTxReference)

	log.Info("New block chain.")
	newChain, _ := New(chainStore, params, arbiters.State, nil)
	originLedger = DefaultLedger
	DefaultLedger = &Ledger{
		Blockchain:  newChain,
		Store:       chainStore,
		Arbitrators: arbiters,
	}

	log.Info("Init check point.")
	newChain.InitCheckpoint(nil, nil, nil)

	return newChain
}

func newBlock() (*types.Block, *payload.Confirm) {
	block := &types.Block{}
	file, _ := os.OpenFile(path.Join(test.DataDir, incomingBlock),
		os.O_RDONLY, 0400)
	block.Deserialize(file)
	file.Close()

	confirm := &payload.Confirm{}
	file, _ = os.OpenFile(path.Join(test.DataDir, incomingConfirm),
		os.O_RDONLY, 0400)
	confirm.Deserialize(file)
	file.Close()

	return block, confirm
}

func rollbackBlockNode(ffldb IFFLDBChainStore, header *types.Header) error {
	return ffldb.Update(func(dbTx database.Tx) error {
		err := DBRemoveBlockNode(dbTx, header)
		if err != nil {
			return err
		}
		return nil
	})
}

func rollbackTo(targetHeight uint32, params *config.Params) {
	chainStore, err := NewChainStore(test.DataPath, params.GenesisBlock)
	if err != nil {
		fmt.Println("create chain store failed, ", err)
	}
	defer chainStore.Close()

	chain, err := New(chainStore, params, nil, nil)
	if err != nil {
		fmt.Println("create blockchain failed, ", err)
	}
	if err := chain.Init(nil); err != nil {
		fmt.Println("initialize index manager failed", err)
	}
	if err := chain.InitFFLDBFromChainStore(nil, nil, nil, false); err != nil {
		fmt.Println("initialize ffldb from chain store failed, ", err)
	}

	store := chain.db.(*ChainStore)
	currentHeight := uint32(len(chain.Nodes)) - 1
	for i := currentHeight; i > targetHeight; i-- {
		blockHash := chain.Nodes[i].Hash
		block, _ := store.fflDB.GetBlock(*blockHash)
		blockNode := NewBlockNode(&block.Header, blockHash)
		confirm, _ := store.GetConfirm(*blockHash)
		if parentNode, ok := chain.LookupNodeInIndex(blockNode.ParentHash); ok {
			blockNode.Parent = parentNode
		}
		err := chain.db.RollbackBlock(block.Block, blockNode, confirm, CalcPastMedianTime(blockNode.Parent))
		if err != nil {
			fmt.Println("roll back block failed, ", err)
		}
		chain.SetTip(blockNode.Parent)
	}
}

var (
	randomBlockHashesRaw = map[uint32]string{
		98216:  "43b15b00c4b54bfd7024cdecb70e165984d5a6b90a5f2e26919a8535b4513f0f",
		245940: "3a3d6032d8faa6abe9a5e5459b703b4948f6cf1cdccc7ed5cbed18230f4db2e4",
		220867: "b7a2f740dac3f4ddb4d9703738cdbfc89bea8f5f4a625e266ae7cecaec9991d9",
		173427: "64295309328626ff10d24ad2747a167b404f1eea4565c41b6839719a997a201a",
		228202: "3ba73e04bb07c780a780e893693288dbe0ffc5dfe9ac17176222f8f21f682a18",
		6678:   "23e7e626838e4104bbdd3d0f141d7dd99296f9c11e5fa073f04a339dd72ad535",
		93739:  "5acb85bd5d87793054d0ad24161ace5f42ed941dc9772c9751455e1d2ab59804",
		204458: "11b3f438b1b18735710e36fcfada4488935f6644f4a40e999fba9f2c9f2899c8",
		235438: "d4f5859ea0919ae9ad661419b560fb081451adea02ab6ea40c62505e129a13d2",
		71509:  "50c83b40ca9734f509ae0d56cc1702c00288aad4a1f86f1af2a609fb19046079",
		251561: "ffa4a29dc21376daae19e507b3a95d33d400ff984b084b5920cb9ba58ca043bc",
		79196:  "93f36ee4ddef5cbc7235bd4a07399fbae74d83329f0563a1cd71c99c7e6c30e6",
		149234: "70daa911cd310e9f03313dd704bb73460300f10a47c50ab2757556dd373f354c",
		105042: "5efd92825a97a8819739750c34d401c9dd370bf55cfc07be492124e9ef0499be",
		252607: "ecb968a42b25ba85dbe416846a1b31ccef521e88eae92849f250eeca199de14a",
		56343:  "3bf4f44ff5ebabb6f6b111fe29674600bac061b4e93553fde03dd12c1d389959",
		43137:  "281b79b6d309f8207b7ea1faf1aea93e53881f3754d22d4faa0010069a004052",
		97516:  "31867898974a0d65f91f9fde5459bf35117f42b65fa8ea03611c355971e55061",
		111333: "0c8ec84e21a3b35162ce015ec1bf2249ff2aeda296e735213cf1ca570470d177",
		225438: "5bd8ba5d391a329892393c2b72824f5c1c4a7e8700f18ca430ec6b00c59dffc3",
		211738: "cd91f9572a8850ac3a257f9d143c14ec0e20cf0c0dbd528b76392a1fc35e9977",
		129629: "4546d369c5abbfc8f0965497083d994db4d906b3d08df636fca110f769cedd3e",
		5470:   "06481181cc1a3c7c714f91167ba158c0fdd43352b59493a2b17e6a28e9bb9523",
		2876:   "d9fb16a53f9d3bf49df972be4a4b0b175271f0b5e23bfbf5acf3b8f55edda4c5",
		196002: "25a3417c1a48a087b25e61ac360cf635a878a44384b686733518a05968fab5b5",
		31393:  "bcfade20c8416c74fad39746ac3f09905d2bd594effb49a74b623f988aa8be79",
		79443:  "238b8abe42dd0ca5092d02e02717ad2322491d9f3f4376ef30155664f7998906",
		93852:  "e9ded3f9f4c0bebe068f35700be4f24f8d82a7b45c7d9340a577f43b824f0f30",
		51396:  "c4ed8e7efa808fede48e8f48233e48bfb84bb57c40d384816f912fa0c951c068",
		177159: "70d36fbcc23a162181d5f3ae2ee75fbde29e9c12398dbe4861f2ff42657921b5",
		35226:  "499f073ce712c0f468438f0a310206708a14ea4596d381b3358b367ee88e819b",
		142270: "4eb9d4d5fecf93139f43e90c014f0ab1848445c93e62648b82b755a41f1f6da7",
		158919: "58f697fee27703986742945c22c5e2b3fc63955d90a5f53efee2a2a537001f39",
		201458: "9bbd298588a7b5f79bd96c38abc46a2dfc34745dad1444a7d727c6f479e0dc08",
		126596: "5b70f05a5c5249abe38da3c52d0a8f30913eb19977afc3f66dc864eaef3548d7",
		225902: "41320a089d9d30635578e8c1c94f9981b4452e403b5a4913afb7d8d035c054f4",
		146389: "1b90d5b86e0f5691aa184f4d002b621fabdc40cf3987694a59c06a8f58e347a0",
		235783: "3c85a9e1fcce3af1bd5261e9318086ca21ad3d261f14a8e06a15cdf952e7d7ee",
		158503: "443b296b1ca609a54225dbdae1facdaa5b84d3589bb187b99a2c2f5fe5950c57",
		170622: "bc49601f93dc92cbc27c407aca912a57fcd7e51fb9a949806dde3a240fb0ebb1",
		170371: "f42ba33c1a5243d2ce6eb091de1f5df0d8247ecd772634bc2d1c1cfe05cf2889",
		141502: "0a72af7e84a4066cab307069fcb2775d15d0928f3e7decff40e9751c22176b05",
		158487: "a418ad6f72b9c23f20fae9313b99ccac6e5ffbf73d9be6948cfdc6127bb532e4",
		14191:  "7d1ee1c0636ad08ee5922445915670e1d0a5ed0bcbb4ba68cf78a169a8f89ad1",
		58326:  "4ce2183db3a15353f7487717d19c9ccf9646f6634f1933a575a660330caaf206",
		183565: "e6975c79dfc9d114dd0ba42859f0dcd118197ceb316cf28983ffb4d620e0e560",
		17876:  "51a9813f251d87f47778b23e59f55427bfa566a8a78ad67527d576c24d20fc42",
		177246: "0be06b7afcede4088f17f977eacc12e750b9d8df93a3da9dc4ba3632b8a8ea3a",
		206401: "720d05ff571a1b7963fa40b306dd0f048944d9bebe53416629c7c068dfdc8f8d",
		63371:  "2ce63322e1493969ad015355c6dddd4012f7feeab0c2dab7f49863b9885b00c7",
		239408: "1d5f3c3a73c66cc7a56d37f37804f75988e048e0b7ed2689ca679e30cbb563c5",
		15719:  "7c1f8501613f02488aab3d114611a9b81c5126f18b8c1a8353de6dfa9644816c",
		128854: "2fc23897ce39b0f7c96afa57fc5ca292e59b129aec17bec20302013a4aed0bba",
		141677: "200160c5ccf26352310d6fcaafb55c70642cf07361c3026b7f46a393f8f8bc46",
		111563: "08704740b67684bd2a438615cd629f618276db3e7db63f199290088534f37961",
		113368: "0109734afe4e1da8f695adf78a1db13ae697994f2dcda9aeec41b078c8d84a11",
		196997: "0153ff418b7da53cbd8c04e9c5c83ea47c0c43c9da7cc232902f8aa2dbd072c9",
		230143: "84593e2fbde2c53f77f78122f55d9ee96185de58c5c067447181a3028b0c41a8",
		20910:  "ce811c87b140ed12282569ebeda70e7fa6b7a19ddfa8d735253dd1a4cc617e47",
		240631: "7ec98518a0ec13fd2210781b8156ee0c56c64f978b37c466d2193cdaaca2cf2b",
		11468:  "726193c84aa63e10950b186978b7364930ae86697261aee8e29c3cd5d57198dd",
		155809: "650fc8293c54e8b2efc20b50c896016ce0bc12d428f993ed681bf94b46df6d30",
		198571: "8c0877a320eabc66dd305910cdf1e0145c1680da329381720f8a65daeae1ed65",
		154884: "b6546c96b7afc23d623df40b9781c2cacdba439ef203e29f8490f5184cecc11e",
		39492:  "bb38fc4f9da39bed0743e8208387d7b13578c0e88d8bb7e2a2382887aec9db7b",
		66158:  "4e806350a2085317a0d56fb4441e62569aefceecb4d6d8237cb1f058520650fc",
		141093: "af44fe728e6b6218241179630af57ca56b1c213c85eeb41fef029984209879e7",
		293:    "15ab2d6a9774ec07d3241b32532c3ab91a02e63b09f1361c405447ba8f713f84",
		10610:  "91dc0f7e0d919a8495464ad3036bd21ddfd23c7f3d41772f6f45edb57fe62d99",
		170083: "01419a703c1e2ee41abf8b219fd446d662e635ea5e5b001397bf28c8a2c0c91b",
		224828: "efb03cf596895626187264107a810f9161f0f3201a5019fdd6c571de4466988c",
		128228: "5cc92033dd3cee5ab738aafd91553cd89c940218d180496394bcdea077fa2787",
		138485: "c87aef26b2506ddd1be8909e2a7abbe05851bb6925a121c8e03be4bd2ad56a41",
		92507:  "cf8eb4513742eb34c0ed854b3e5f0017a46445ef1e5215ccdb471c84443b32ae",
		56420:  "0eede0015e1a6fa62d08e52ff0a0075b9eee7ffc95806e857a70ed92cb795ab2",
		194538: "275f57261209109afd09ca87109ae5bb4be7c841cfc7c58f1744b8874c5bd8ad",
		164789: "539318d6173e7ac8089e8e59e7d0f32f9ed557dbcf169fa992cc4a8c08c37578",
		235652: "c96395a5061031dee0fcceb804c7e4368505b2ab64d47281527eccddb72d9763",
		219517: "a0659c56201bf5e91c9ab99d934fcbc3a64b9db4e5eb81cdeb59eb69731a9c6f",
		107461: "8c76ea9dc05a03450c29bba35e2e42f97a3ebd69b6c67ad23b75ed47d2474dc9",
		172890: "b284f95dd9124d65ccfe13face4841680344afc7dc79b842c287a75ae16e4ea1",
		170843: "235270cec3cdf977bdbfa7a5275098a03183ce83ed802ba2dbe5ba4eaf60f09a",
		190124: "810c59d4c4fff381c99086f71340ef262d2372ee60c08c2cf9117be624e59a87",
		213148: "19695ae535766416b5dd93565e050a335c8d3e67d8dbcb9bc3641465fecbb401",
		222854: "f5c2dbe2e0305f9911e3391358676b9baf4f422784c2d16c7dfb8fbb57bc3889",
		135434: "e14c293585204d13da5dea7d8aa6474dc74b0e6595d6e6179ebc6e3b355a3b44",
		188045: "f58777cdd4c860b28f5dbb76c5ebbfe072069ee08aba53aa62834718a135d989",
		210060: "74638880290e29d26fb8b4d9aaf1e0471c208cda6f95375e60bcade7b09490e2",
		20179:  "1cdd35930986d8d5099c0cc508f2a91b914b65db19ab7442082aaa082f6d56c1",
		97403:  "deaf9a09da754a01e935b9ec16f6d5d782fbc553cd399333fee970d6a210183b",
		244647: "d2dbab8fc36c0d27609b89e91047545dadeb290e68287c03c7b109e50831519a",
		2899:   "51257841839a5cf7236abe81589a9f33541d9ed56f8e3a9c6f4b3ad87bb6e7f9",
		132722: "aa51b0c54ee7ce5678f4ad6d5d76f4074ca3cd7796e057a9734c87ee5de6d0c5",
		218674: "9fb7cd7f2b2ab2e622d21e9a079597a18eb190640b6b1dcec861c3d3dfdd3063",
		15625:  "7103c05d547db3740946e388b4618a561e5b25b21de3632c92b65293be7055ee",
		55578:  "446ab4f06b2d71be3e1cc0ed42e340b5c1037edc27e83137c6c6a0e7d4dac463",
		5082:   "820f3e75c9e2a76c868e63969596c32fd7216304b8ad0e390d3d6bb773685f5e",
		248368: "73a7cf3ef56fecaad5d61abe4be1fbd9cbf2e080e7b2e5baa887c53d3c6779d7",
		111417: "75d47f3991e1f831dd084e12c3cc62d294e3f7356ff4f852f64d8ff8701d4ed9",
		50571:  "9a48b071ba7892fde55e343eb68b2076807b45c43e7e5596933a42043af774a0",
	}

	continuousBlockHashesRaw = map[uint32]string{
		100000: "f541a4b7733cf89b1c1f9320eaafd5475a1234ba1d7136e40bbf41c5fff5df46",
		100001: "cdc5efa1de01578065643cfaff64cbc2d9ec4e36d14b1f0d6e77b6191e4c8460",
		100002: "2199ac580c5408b0fb8bf7e2dd4c0034d34607629de21c7ed0ea18424b747134",
		100003: "4599417929f0b79e1c3a52a8d125b25f530bd7ab872f3f4ce851490c33cd6372",
		100004: "3c6eea08ee3d13785ba7c7eaae29b0dbc986271208d947a40a46b77ec934e0a0",
		100005: "61d8feb7fe8918b14881daf9988787886c6d7690a702a2cc5cdb65b9ccb53956",
		100006: "32a0567185d3265d4ed9bb927cf57d116b57cbd6b956af47648df444dd60a0ec",
		100007: "c279a80e70483b7143de24110e6cdb33f902f94ab268558b885c0979e02e85e4",
		100008: "eb478b069d958c4d8f578d054a765bd69b38beddd5add4d2c06ccb3fa8cf4459",
		100009: "104481e531fce34129e987269514146d5f14808ca010878212f50f3cc06fb291",
		100010: "a17b295549e02c1c19080dceb223386d85106c1a2c0052da2dbfd1252586c48a",
		100011: "325e074162129422f59491f22f30a7d3271c629406e5946bb588e42d5c70195f",
		100012: "60061a3c3f24cba1844cfa67a278e4b2b59bad090893f44ab83743e41124fe76",
		100013: "b1145a1c6fe3e97f13170a0e5f42f7f704e3086c2bc26c5ef4190e68ec951a18",
		100014: "edc3b0cd857a1b17bb65a6963e55786b6186313f0d68dac12ddd8d32f30eff26",
		100015: "ac791796684420e92186e9e86458e4ca0230f935188cf60accf65f3aaf6f012e",
		100016: "e101b59bd83bce5647b9d5434114d7a3c3d8951f7cd1b4609f4a872eb58fe8bf",
		100017: "cdfcb1bf5196c9e46d838ce59305656f0c532a540d838122bd1526335cdbcc08",
		100018: "242603f12566ae9246006aaaa1ec20fbc2df3f1e82c559b4c38cebb6168b8cea",
		100019: "0824ac0e7e1ef31abbce34698d0f8740b5ec237eba3af9b060e3622e5731689a",
		100020: "dc085d4d845fb0052e609656480622c6652ed1f50f146b2bc6a7f8068c27778f",
		100021: "b7c08fde7addfea14acfb3d449f3b52f2ee8b5330aea3aa0b7b9ef6753295244",
		100022: "39713eb77b388c1f8911b653cfae7a6b8d01a04782d44df30f46c10b1d643883",
		100023: "c6181cab9de04a1750c776b6c3af960fd378511263c6bce5f6ed8f2f32fc660e",
		100024: "49a9907b453d6ab3c73dd73bb003050db3aec1d694b66346c9b7b2d3c6b34431",
		100025: "2cd39ed483d14fbea354ed5813641972f6b9c7978ef00849cd026886ccc81abb",
		100026: "8348a897c7a56d5d7316db7c4ee14fa4764a410720184eec040d8d8a76ee00e9",
		100027: "a1650906867e50f87582b081c8b17bb8563f999f539230ad814a361cffb19eef",
		100028: "f25653fb687bdb3c0a341b39958d9d517b927725b7412d0db3c49e532ec0385a",
		100029: "343a86db7a32f66df26b1413271b5f19f9b02d5d1157bfce310a5906f04c0d6b",
		100030: "b6d0198fd3da8dd86a3c26d2a880da238541d375a4efadbf554b2db94b5b1f58",
		100031: "193cf499c06475981229ab5a7142aebac9772e69d61ef4f17b1ab78f154322b8",
		100032: "a63c7cf9b63381b8d86bb2c8803e4f36eab485426ae29a12d3618fdc4188caa9",
		100033: "9015d88dd36ee99ffd6a08cb1b6e711995851f0d12c8006f97ec7d4a3f17e3cb",
		100034: "dbdb2901fc73379a619148dfc2d01f43532e33f806852d0d53702132c61400f0",
		100035: "4b05e53a03e4009b430f628bc7d9553d7c22f510526565516ce80706747af7cd",
		100036: "44712615c922c95fb607df30b7c9ac3e32bc36140e39dda8c83c812190ac2389",
		100037: "67172df60ba3341533651f98f45407161fd3293b6b19ac684a96ffd86e5e1aed",
		100038: "da4207f99cbdcf403696409a50a758b08601eef1882fa74ccf75be2a5330e63f",
		100039: "daca916c41335f6bd4f0c44b37f3cd70c398326c67c2dd7865e65d421149127d",
		100040: "0d3ce834919917fa1a30960286282a48e2f2ddcca72fbd3419a6e65d36a458f3",
		100041: "76460ea9ae07b514377253a227f69ba7196d710cf4b246b843cd91c00f9c5700",
		100042: "22147836e7911a5bfb423f47044aa11addecb4a573ca1c95d28bd25cb5fc9115",
		100043: "7653d54c284cc388361755b0de0cf8759ec1de9bb2136cfe99e594fabcced9e7",
		100044: "4db821de5dab797e02e7f810f8e2a8df62c6a3b21af9a0e73e51bc3c338f2edb",
		100045: "afc45c219eab41dcc5b2bcb5f4b053d6d0eb9926c001e30736b35731888d5c64",
		100046: "e5c0b622306ea6ada0aae372284d49b9247b944c4fa506c355118ea7f023a5de",
		100047: "1f261f52e635f7afe7768965e6d58c8920b35f7ae77361feb80ce8a2977448ac",
		100048: "53ad0e45ebfa8acacddc6aac215e26d6bf1ee9db8b159be54b7e76b5aee4fca5",
		100049: "b560cef92aa92a104e81a2bd9b5d0535cc339bbad4c2f73ea402e85f593e0e11",
		100050: "9d6a5c1271bf148a5097d92b5627edb2838675e9b977328dbdbbc938d17561ce",
		100051: "0c22a3ac780598f477c7a9b616227c6409729ee0ef98b69c3cf00c57f4294c6b",
		100052: "a80b88e1e01200b5de74db353e795df9d4894f970bbb4267b4c51d5a3b261c5b",
		100053: "3d201307820a60aed9420e02bb29eaa60a70f4358359a6b95e2c2c01a4f239fd",
		100054: "accc17ec6c57fc1ae57b31047ca0cf185e70faffdf0ebbaa27f563e5e30d77cc",
		100055: "a34b1ffb51a6af135871d45160e99498835b19dbd1f5133f6707bab2139f2815",
		100056: "6207b7ff0dbcb85e0f6d6fee4a3e96c559f3305ce41d0b8f0ee3b7efc456589b",
		100057: "c68678112c0fbe3e58a6df0c0f12c5c03ac27dfa5a2c4ee92d46c750878505e0",
		100058: "484b2effdf0e0756bc6cc0ae1f1d7a7ba48236d6a7dd1d088589c1d3e52ea6a3",
		100059: "de201633f9be92707f5545872912d0a7b1237be11083871721ddb199f6161956",
		100060: "8ec7ab7bf1817c9784bbe990a968ed8691db495392a9f751fa4d400decca9657",
		100061: "889f3d0a9fffb2dc0931ff3f056b002c82e12204add75e84d2278848a567b46c",
		100062: "06cc1af7ac1d6fbfa6f471a677f4c722ad1de9cadd2e4f3f7116c2091cc0a923",
		100063: "a98d8ce1450dcc3ef19812bf1b227b4a168c71101eece7ffca3f6c5ffcf593b4",
		100064: "4b5379821833cbb1d8ef3cd6fc03635c871dce7dcd79b4e3565a57d0eb687fcb",
		100065: "fc5b0908809d7ac40facc2bb9eb41416faf6c9f7d3a1e253fa39ba07b8be9273",
		100066: "0a38b2346b781a4efe067b2e6545e115ef84ec235139a0c4f3544c8c4adb3fa5",
		100067: "61cac9ef53500ed4687ef4f9e7d0195f08bb740f4a316972a4c6df4cf332ebfc",
		100068: "12c82aff3c38748711bc8bd1d7135dcf28da5ef19b9df9d1dd1fc5b3c3c8345e",
		100069: "9b02d4c21015cbb1e4c2ffed7712b42c6f77d1e2d60cc2a88c95c2e33d6a6870",
		100070: "c75f07bcfcb511534887fcd035ebf7f3aa1dc782c51c01bbfdc239ed8edc1af0",
		100071: "a7c8e08785baae13694c65a58c693b4b99be72e866e9fb4cf5eac0153a62fd71",
		100072: "abc53e5092786ec7fc597372fd8d3d65e43982d96f00f3e9f2dc613403f83234",
		100073: "11941eb32b95fa59ee0271408ca55da307855c3eae41df5dfeeec3046cc03f88",
		100074: "a358d0e9325031bfc22cb4522429ef77c9daadd440f8fd7a8b4848ed3e378097",
		100075: "94475fdcde30a42fab10fc5c975004d20abfe17a56670496d7e7cbac5a94c0d8",
		100076: "f76d99b6f453821c18bb80a7ca89083fef881976f0b6a878424129202690a269",
		100077: "1db711accc8b618d9cd6f22639bee327d7216bdbf41d19aa5cdb88e6d8adebf2",
		100078: "37ad86e2bf3bd388254d38b5509dd24b9d4600d250eb6853c847cfd5261bae60",
		100079: "914474a378143ef890eaea1f31b27a870739fd7ee4355999fdd06c552a35d34e",
		100080: "eb180005ee0e0a633ff4681584871186cd19c1f8974b2f32cc682b199d96ae6a",
		100081: "c4aaa55aa1134dba629f841eff05b388f0d9f8c9894e425d28f1ed3cc078c253",
		100082: "1e90cabc2b6fa9eebfae39290cdb918124276f32eb87d11128f9f2788c76059f",
		100083: "07cf6dd2878704556a17bba5ab5ab005a1a357f0ebe95bd069e67a774d9a6ad1",
		100084: "845609cdeadcc49580cb022c577f55eabddd8b9ae167fc484a61eda6fd1b1dbd",
		100085: "b2a7e49f836c69f49d777037810e37ed47d6b12e06cd5375e2a888dbf35568a1",
		100086: "05426f7dcaa6735b40edb53095a4cb3aa5139104f06a0a157afa5046c0c9e940",
		100087: "939c90cbb5b149300fe5d9a94bf0018efd26139b4bf232da443ccfa18d6946e2",
		100088: "1917019e217e7f7c8d834a29f880c556a229654f8bed2c8e826024ed802b7fa8",
		100089: "9626c8be97c934aa0a0ed02d59d701c3dbcd76d89f329deb6307be20747d5ee4",
		100090: "e4e3a4b9d0a058fedf80ee612e6dfe416c6f552b2c2d929866669b0bcc943a94",
		100091: "9160273c3a21d8dd5fd3a54b49920de40c39dc117613b87cdd689863a867ee48",
		100092: "bfc0e8c2fcf7d8548009fe63a0d44eab358c399d8fa6e8d4ad9b8cb2043ed368",
		100093: "98f62e6ea3b6e3698d598cdbc279944de2aed4f4307ec71066a18c09688dbb26",
		100094: "b824c4ce8d5fc17e4c57ed6877d4a2bef30c32ffa015069a8b5cd70d38d3ef7d",
		100095: "f86a2cc04184cbbb6eb8bd83b90695dd5a81bec5b7e9554f01a3eb97326958ff",
		100096: "a54d0b76e5e6b24f0d536ede81214a5eed4e423f54f30ec3d45f1a6188f0430e",
		100097: "b4f7a932051c024c22126c88966827dec0547842376fb0521a5bf718c380a30c",
		100098: "43c1f06d3737d7fa833caa0fd08d888adbe1c210bc82c2b7858d27503a016c9b",
		100099: "1cf6bdb47a747392f5342519ae94e7aee1895c2797ad9fd0af70c08182ff22c6",
		100100: "5e9ae637e9ac2274be1617eb5c74978490989eb57fc93f417c298d6c178f8542",
		100101: "36cd05864c56b2ee0e18e4fd1bded3aa1375e7931470435eb471a3508e155f3a",
		100102: "7e49944e49554034756cc3cb2fe8e09714d38b38ddff2a35112da315ca8d5ba1",
		100103: "bdd43f1d576d72c28ac98415fb3ef625edccf511894d6dcd0f5e3b92d21fda78",
		100104: "92885f4cb7c06059427e30a88be0b5e5749affc5bf41b372e8cf4c42bf5f76a4",
		100105: "6ee21ce42fc7d2ab84473083c1c3446b16e8171f7068b929320dad60cc56262a",
		100106: "26fbfe067d45933ade590760735748c2a6bc80a612fb600512a95ae6add3b6d2",
		100107: "64c50db5f80a54f8f4402ac3730e116dcdd47832cfe98009802b8ad0ce7cf814",
		100108: "dd7ba1ff2e04136a1f05717eec4e98e2aaeb3ff7d133b807fe58dc4e5dfc22b9",
		100109: "cc4c74206070e3e25a98344be40d8c52551627d1c26587bf22cb51ddfd1aacbe",
		100110: "edbc93d4ad61c1bdb15e41893e45b933aa621349f2d1e2ba6c8f9d6ec2f09f4a",
		100111: "3c51aa3393c2988da3a65fbeaf3906f2e3714268c9fe8a81f10602f798e2f405",
		100112: "4266b9ed971fc864f9707b833c78a01ed41c919b1e99763c596767c36454c91e",
		100113: "aaaf96964a5e281a7278c28be1f2fa7641a10c6e189b453a222aaeeed2c41c69",
		100114: "fb3a99163249179e76247629073798c45107ad9401977aa0f0b0245e619f861d",
		100115: "495ab43bc56c671edabb573cb71584f2dc08cd619bdb8073eb9db1be55e8c528",
		100116: "104c8ac15cfb9c2ceccc67d304d1e67545bb4fccf0ec5a0f0636ded07938398b",
		100117: "ea5f84fce1a76e71c81a6bd2ef01bed20adf62cb6db9f34a6c480fed1fef5af5",
		100118: "51e0786dcb69bdc1de750449e20fe1a6afbf98796025ad99d7d823b7371ef1a3",
		100119: "bb88cfdeb9234f6108d2ebfe66d108cf83cee8efd7dd872a0e6946b5c921fa58",
		100120: "6773430f9e81ca804f6f5c1ea3757938c149a346743119f77f865a1b99be8602",
		100121: "feafbebab20b0aa3f9af60557497479e30eaf1d53177618f55cd1eb51ac980de",
		100122: "77bebec19430b358aee7d93f061ad2602c3909796397d146b7c750e33a33e65d",
		100123: "b3dc014b75f0bff93502cc8830c15f327f698b0be8766cdb08c913d51edff6ed",
		100124: "e8907bf4f2fe030ad2d41f585749f3aff040d6d804eac652b674c069a3c6c46d",
		100125: "94a2a0cbbed05065dd67bd17e3d98b19a04a796e96b75316b617f657bd7c4258",
		100126: "c9d7ec63bbf8695c13e4dd290be0a4ad2a2178d415f85148dc36ee14834cdf18",
		100127: "7c63403f6d0aa8347a8d4632a90c40be618b0306d8c6078e47fd114e40f44759",
		100128: "eaa4b32f5837ef2b1a55823ab3960424bf59f94d47908656d667c26db03263af",
		100129: "aee6a6a3b3259f719228fee03cfec4e7983484fb26270972a072ac90da567c8b",
		100130: "5e5bde3a28b08abfda4d9d4b9fcc14a09b0e50a662a451fac014204363e92b53",
		100131: "a530d291067538b6933a75cfad789116d4d862c9d0bc319daeb4a5e883904150",
		100132: "0d88fa293e8f9883b1d3f18010d774baf6a94b2464b62d40a33082f00b46a22a",
		100133: "8c9d282ad575563a0566573580b144482c77e6151cfb2e6e274327fd355fa9e8",
		100134: "75be1ce57792041c13d8022609e17faf3752c028b788a46ee74de94e1a7d38f6",
		100135: "432bd9b9780a6c6b06e4ab52121580c34bfa46c6bcc3c5554a27650016089543",
		100136: "00193acdd8c01eef54b88c9f8877e05c9a8af270df8f5367f240b225f9b06fe3",
		100137: "37aa81b8bc9d088ce226ad199a77a1544b3489731c0f0dc519cb034beab46b66",
		100138: "b8b633f41e9eb7a41109c765b1ea96e7c94968ef98df4da2c4dcef2201b58f08",
		100139: "fbc47b8b495c4ad9312caf0a1e28fbc3c554f6bef609df1b48f042015cb56e42",
		100140: "e6189f103e530679034b5babeeb0ca562866a3228451083676df922f71fc11bd",
		100141: "4cce95e638cf0e8209e78dd8bef9a1f3b8c6cc7d985a329a0ad55b64c33d7f0b",
		100142: "d62a9965d765a5fab4909fd25ac2a80c4ea88d9537eeaa9fa69d423ccd7bc523",
		100143: "c5679f4be6dc0541888ce0bea1f34382c0c3e894ccf18413dbf2fc08fbacbca5",
		100144: "15a0d83bfbd6fd62fb3454a5958fb498716e1e62beb96545bc1db4aa50670dd0",
		100145: "fd1a257e12549ff42b5c876ba5a2fcc797bacb63cbdad0df1f3607c363526cab",
		100146: "7db787201c65b2cd414c38bf711e7d2086f7e7b6a93f5c1873d5c2f37ded5465",
		100147: "424da913c26f96dd4b68b69e4330d2014e0605c43cacc96a7e36f734df650941",
		100148: "c8e356171558c008e1116cc3887729407c8b74ee4226345a5326182cea7028e7",
		100149: "61f613ba3887c74b6d68731908e873453a8b67db110faddb2f4010b4feb415f9",
		100150: "664758fb298f5d99603e4bb08ceb142291d243ec6d16f64fec5d094c2900c83b",
		100151: "ab81115d0f87973b24a632e085d04747cbbc207529e17c037a34a2cde3b085a5",
		100152: "ab1ee768c8acaabf294ec9d9d7780b5be4ed33373b534c22bada8e6ec1556180",
		100153: "d29cee745659fac4565f690f32ef42ba8566ed13ce3f23045be653d39431c712",
		100154: "07454628a60edd9400c4e7f8da12852f7adba708dc4d5ba563e7b58b085ef61a",
		100155: "8145e8ade36558543b3ff71a2b5bc323a92c3ccefd7153e4b520aa140cab67a2",
		100156: "8689f6d5a367652d6b394145540f52dd3b91a5282b469c66b869ce1ec98c3d77",
		100157: "8c75cb671ed6257e8d9bfcb5eb97a0bfcb5a6bb755d526fa01476a871b41d3d0",
		100158: "2c902dba7799ab092c00bb5b7a11c09b536116987858a96145cbc8aad9703fbb",
		100159: "9b0b2631cafac597aa3057799a88c7a9108695df2bad6a0fc9bd493cc4091ae1",
		100160: "28b1e5cf891b297c00a99c90eab4f1f44233d14331a1a139f19240565816dc5c",
		100161: "39d9f8a13371ac962d7f284cc9e94fe26d43b5d5792d720d946d115b7216d9d0",
		100162: "a408ae7e4b2c221eff16a3f48cc36ff303aec94a713a202296847107b4f9d509",
		100163: "57b142cf871fbde2c69146483b3fb2e620aef0c0a536371bb456a486c9f4c05a",
		100164: "12030df08c2b7c4b66ad88f413f00f774a4bd7c00310f9412f5b6814d5f7e4b8",
		100165: "5d1389b61a96a280c34a1d1659fea3a1e551f9b298f24596a6d98cb6e29f4b6c",
		100166: "4552754b93789ae0570972585290d7adf59e88c67e238a1a4df9c730c634bfe9",
		100167: "08594187eb56ff51db5a9d24233edf63a20b1be0458c44b314ea048b7d0b40ef",
		100168: "905a09d5b8d21e2eb4a1c516acfc40de104523899eecf08e7ebac13557c5d2b5",
		100169: "3b4952f15ec40cb41ca47a5e55103850f8f8348c5d6e58229c28ee9868f3a89f",
		100170: "b48c68434fd2b1079a1fc8ebc9889d6300e5f5c27d2f402d96709f3ce9769ef3",
		100171: "eac138ecccfb08ab3759217e4daee45f5f8215ebef1c10f01053f7d1d58ae1cf",
		100172: "b360b83c581a74d4b3cb466751839dbc9531a107c5c64619f2d49f5b9dab5541",
		100173: "f4e35b26839339321babae39fe7380d3f5a8b4e1458687f78fc8159e5ba55303",
		100174: "59373c0d9890e99e45240891e5fe5708274defea90bc0e6ad61b5e80463c37e1",
		100175: "f3a3cd4f2bc1b0c39ae6b2c2bcfc6e5372a7bc9e31398ea6399aadd927b42f48",
		100176: "e247eea570fedd9fa2ff1ea1321d190ee8e7710f2e3b7a96d59d7dd94ff2c161",
		100177: "870e3d7b6c1abe41376619fdbfd5f7580d797a7b7b7b4b36a2fd2f2c069d3bfd",
		100178: "3b128bfeb5be6a7b26a7cbfe0b8da1799a48d836c8928bdbe35d64bcc71dced8",
		100179: "4174ea841802f289f1c5aea71193994aabcb8b98831035c224a8c31f188a54ba",
		100180: "5eed787e46e3f339fc332dab681fac87b224694039716c7de796c8757dbda896",
		100181: "fb45fae133fd2d91aaeaac3b1730f3e561b6b396a5fc3a9479d30dc9396eb7ba",
		100182: "d2cff5f42ce8594376f0b8f50ed3dcdf52a724dd03bca48465ba0d57e0c2a8a6",
		100183: "a3035b9b6cdb90d7f3e87e2d97cbe837171901c40dc5f4dec827dde48388a03e",
		100184: "e53ab2e93043e706afc2f9a2991148899affc4eaf2afb3b9b9d22c60a2c51c70",
		100185: "93aff7d5faa007a1d7501f79ebef4d67145efef55d38039e26a40f433e857e76",
		100186: "b7f0c9fb67bd3fc322a1eaca3e5e1e3f01c0a00317eddd963fd244cf9c1e4678",
		100187: "cce24641cd43c5f996cd0dd5b3621142757767fac6d684ec3af0e3217bb17af1",
		100188: "a36fad1b0a21ace63009f14faddeba1c30938ebe4ea840fe0a61ea03a5afd046",
		100189: "d5ee540c7469a49ab800d95bfbc7b85b29cf9dd63683e853b1abd237b8bd875c",
		100190: "dbc8c9b96c2871a9d52d150695e31fba064947c4350c4616a027980db9734700",
		100191: "dd03d97d5b91d72a2bd5221430b258ece00dc17213e1441c2c49c4f9ae9d754b",
		100192: "924d74400b35f5e7f99ae1963042288a9fbb16805700792f863300e8bee82f31",
		100193: "4e433c2258c5f194c900b31d76052c1d9f3b8cd91dda7e03d81328f4001d8e12",
		100194: "09b666d408f89cd6fe378c7e4925d03dc96394d83d3fdfc51ae6c87ee1511fbd",
		100195: "2cc2a1859af54639ff3ecb830bb898837b217ebb59bef803138eaf72157649f1",
		100196: "4b958e6df21c26248711710d776247684cc32515fda2487301ed6dd39c3099d1",
		100197: "1aae3dec8ac20b8a7dd9e4c960aab34ae12dbea2a9c8c586d8b9df9473a8f5ad",
		100198: "0a6c485fdbe37bebf5675396a3903a61dcbe078576ec6856fced5734a5dde290",
		100199: "8cfa374a9650c38adad2d8a4bd2936d10d56fb8f62fc702bb5e7bd8a20e8f033",
		100200: "defeb056e6850b8b430ead1dd1d25bdae39f21aaedbad1b298b4e0484741254b",
		100201: "a02d5c16bbb809953416d40bb3c0d783c42836244badd99cafa4b101e2916f31",
		100202: "0311876dc596e8e47f24e1446d558a31d7bdd1b3976bb9cfe707b879dd29ba43",
		100203: "db8723b7e0ba821db5e8afb049f92d0e6327542f08c35a7aa5885f59d8634537",
		100204: "09d3c7a899368210ada2981478240eb0cf751af06f7b0d6797a5627168d8590e",
		100205: "1bf61cc95d45be243c9d76a53b20f37ca7ed46b30918f2a78c59966d48cff4e9",
		100206: "1e8af96ee6db6060b6dd6004561821e4c037b0dfd8aaee3969d0be2f95eeae50",
		100207: "6f32d986afe5f17943c3935242ad89204ac034cf20cc0cb69554f038a62ea390",
		100208: "6441f99bd992fda68696d5ac19c847c3751ad880c1784785c6602bdb3438e63d",
		100209: "7f9009ca5bce975f967a5afb13410158a5c6cc646aa006c8afc57fc481a2c3c3",
		100210: "f5b77dd182f63ac8b89773ccf5c12c1914c8727adad79587af7ac588cd0bca70",
		100211: "ac899c6855689ba142877d539a7d48e2e5a1e4d7ece47e7ecb15db0a5c5fad08",
		100212: "1f821731ff56beeb697643e30193dda5f681d515386d0e3aba32afa6083418b0",
		100213: "73c9a562574048d7aae8deb2031b4d3da679f46f2e46cef3626f29e0e5505ac0",
		100214: "598a56d434634ddc00a30548253288fea76393ce822c165a3128ff3a685fb867",
		100215: "40762f330c96ec2b342a2e8b51e69058297d16008515e20362019693b09122ed",
		100216: "269f7db033d4f6dc18279f366e4627232bd1510e27a945923591a13c5e19ab43",
		100217: "7e87635a688a2c9c639b2c24fc4ce0626da70602ce2f598d7c3a959d99e4cccd",
		100218: "f063e96e563c987772d927f2b63a1071f2dc5830944a545f5402d55df49d9f18",
		100219: "ff6f047de1d0a9bc83a258dfe492104f01f8910ef82ad67aa905210588f79f1f",
		100220: "6260d3d14c0dd94ec48ed8f10d388c0c11abe7533af375e053d788e2e25defa2",
		100221: "97fa537ddb13230c0e30a98a4f8f562e0c721fdc7dcc620a41a9fb9d12d65af6",
		100222: "b3a66998b654aa6b91734b1cc3d01f329b07fb46d1753900c2129cbade92b9ec",
		100223: "7c1438e659de6ea71944c996ee742aacaddc88d9d968ca943a1f67cae684cbdc",
		100224: "051d6e112c23462df998f82f1f2955c5217c979608cf44eccc52bf5f74686709",
		100225: "c6f9ac152e7402b579cbba6b024ccee3f88ca1723613feab0623d7140d83402c",
		100226: "e22cf5708286582644311b4710c063e7ebc01d09ea691ae841103af025935a26",
		100227: "fb274515f9ebd63f205383182fc85945ba8458f2486f3fc73f37d4438a5b90cf",
		100228: "3ba95f7363dae38eea451992647530a2e57c0596a5ac8ae3d2558ac0539afb1c",
		100229: "a3d65f8dcf0af8d4042897cc3b258068ef9e72bfdecb3ab07657ffb519494539",
		100230: "f4749af4a2e08b0730fd6cc4f5e15308068b31fe6ba4f511998170262eeac66b",
		100231: "29a3cc7b5bd35e53ba2849e011b4728197d6653f2de93bda41b6a896503b9d0d",
		100232: "f2c6d690bf334fdb02d94a41bcc81f3936b68ba201ed0957233d33a01beb835c",
		100233: "90acf3004360a050b7f759879a4dc17670b46e59c5ba77e303dfa94c00cd61c2",
		100234: "7061c57e2439eddd72cf77a4e042745ab72c752ba4e53ee25d686d5ffe9e6a65",
		100235: "a995aa32dc75b6c42413bbe2e351e86111b98e10c4f3aaabdc2706624034d6de",
		100236: "749585580f9c6aad570f63cfd796b4d88de5b1107205602a51b40e2cd89e522d",
		100237: "42a44af3fb5186198719de0f082771cc41c577efc86d118d9d2a7f6e4253280d",
		100238: "b7c9fcdee3711fe04d50326a3b13ad72d99a2372212349cf7029343065eb41f0",
		100239: "371f23da956a5e1c9a56cc8de16d410cbf6e880bff23b38af0c2432f2fa34967",
		100240: "2fe7f4b5f6bbc672dc788e2fa3e4dc6332ed7bccf27b2ffa8572486782f05b5a",
		100241: "3e36859749ee287c913fa7a23ffa17e32ca8da5997fed987cc46afcc5ca99bec",
		100242: "ed6f8a38530b8c65122673315ed4bdeb910d442e0e7e29b8f837afc92eabe8b5",
		100243: "241c5031f1a24d1b614fd726fb4078dcd4168c948eb0bb01c0246b9e150ed5ff",
		100244: "f504765634419fb980a834c446ca416f7542701ac31dab8ed282647d98d9669d",
		100245: "3ae19b82a1a461aa1f8fe0d8f5d0edd20901c0954162ef02663b92f7b8e2a45c",
		100246: "9d0813763f43aa9b9ef0bf02301d858d512dff243bed1df6360ba1176dbd7b62",
		100247: "56d6c33ea71dbe01b9bc00ed111c3c3e460ca53fa19a4bf0b9398a397856c73f",
		100248: "4a180fb6301fc38b54d4c77569ac16c8412ee38087057587b062d0b7ca0cc532",
		100249: "af28b23d6ffcb648cc125d853da8c4dace521ea57d85ae06a85455fd5a38788a",
		100250: "33b5dfedecca5d8a7c7186986a44d882e281c9e7c0177292ad9ee57ca51ebdf0",
		100251: "d37d04f4f7059692016ad460fb4b4633de772c22d383c28a410578a34c2259b3",
		100252: "357a2399044eddf771eff38c575c91b708b9792cca648e94dad03f09bc753903",
		100253: "3b04dbaa1e635f544b66cf82eef4c251774c286196f1673aebc1fcbd7be03f19",
		100254: "8ca750d13c1b95c1b4f54b72bb45257ad647bcb4bdbf228bc357ccb2ee7fb6af",
		100255: "c9e44c14e07a70c59169e56a9c6fcda64b443e9e3c5ce33f9747c5524fb7fe28",
		100256: "f514e54fd423a73fa3d35c589ba3e93f5a544c3e6dc2b36ae083f453f9bf430b",
		100257: "49a7011aa1866fad9a83090916b0899486e961598f73f70d5768cbaf21250f7f",
		100258: "d98481dbf2451a88c59f6bfb88e63dde575656b033f0dd21c633e2c97386b9a6",
		100259: "af3bf0d6e4451783d66830f3cd62a5451732bfc375e41d9c3d29722f9ea68207",
		100260: "09e3c5cdcf2e48f16f756c18f0385250bc05eac0301987a530502382aa108a6e",
		100261: "bb192dc9e0b31d1f4f7cb353b2263245c820a71294112275e7cc8c623e562a44",
		100262: "727bb16ee8e504fb929d0e7011110d20c4c29bf70c79eec3410b89513ca6aded",
		100263: "6e03591b0e74eb0fb12d47428cee2575c528ccbc273d77f019b5403bb2b576eb",
		100264: "4da04d5fcb9a0a9f8a522d0ca49ac22aeedb24f6d1b6b8db7e7137bce8b81124",
		100265: "7f6ef7c88da506e88def289594a2d166363ca60913af0cea3b6d3d5b42407e4d",
		100266: "5b0708c90b0b809f1b8354d662fbaee3f29dfef82014bd2657e98527cff4ed92",
		100267: "74275a4df7bac89e7a5dc86a6aa0fd92190410d547fc72873aa9ff103e213210",
		100268: "b90758cf065de514a4d681c018b47729931cd23d69e422a93ab97dcd9950dd62",
		100269: "13e6b8dd1ccc9dbd88f27d942f2ed52a20d060e6e813658bc4eee77ac6e26911",
		100270: "c301fde07289c1b5a1c93ba7d27957a9e70a402d1ebb0036de10c3ce59f8b0a6",
		100271: "e09ef313e912a47f7851ef43003da82e8cf6f63e9ca03e67242842598035729c",
		100272: "28ca6d8e1a1a3994cbdf003bcf7cbbe10582a608ad81170ab82e8042d66e915a",
		100273: "921804c2ab532cbd04c9fd5e5bf7c7e1511ddfcfb5a509e471c8d5c6f4785ae4",
		100274: "df85434c67e1739f51a5d45937015f3662ee9fbff428ce3e5b749078736a7e53",
		100275: "1267e31240c369d7ba30da774af32030832120a2ac51800d4980f9cc5b6144ad",
		100276: "8960e8a44dfde30e96341a41611b2db5ac861e89bc62794b2cc96bbe89c457dd",
		100277: "69efebf8eb07ea4d27c53f9fc329604001577460786fa998c0d156c4f3d15b49",
		100278: "1159237f6faa3f9467e7beafcc0cd52be1b0aad2ab8de29525c88d7f0a13fa60",
		100279: "04b96db6a564017f9202d56be51771aa16d9f66ad8e0b28c81acb7466818fa6a",
		100280: "a0feb5c14fa2dc5fa43f6e0a99c6f5a768e0f413d8d0401094ed59e6adfc0ad2",
		100281: "a7d81be5f00da65916a3de665be262c127a6e015689793e22bb3b60b9489ea51",
		100282: "28254e0db32f9c2ac4cd8e65f77fd1bb9eaaee62892eaf9107ab21a9f4b1eb9d",
		100283: "1cd22ec095b59b5519c8c01b08d8d415ae4d4ab15fb5bcdedcccbc45c2c04f36",
		100284: "4a6e9efcb049fedfbaa496643b56ac7a85cde1addd80233c184870224270a492",
		100285: "8b350182ac9fb8fce0f70e07ffe519ee9b8b7a2db2fb56f274f71c9cf9ce8adb",
		100286: "9dc25364018d29f93168facf7f270644d55c76fdd6a055cc02d4d25a17648f3d",
		100287: "9a510ce7e8eece27a24b732f7788d754650b416f2b8d5dfd42ce407773176583",
		100288: "4b89362e1cac7f739d1a75440aee2b0aca3b6819b88a6ede283da597c6218246",
		100289: "394bcef6b0e666599e4eec8f0471117b369db64d7d8d9e835dd95e5dabbda604",
		100290: "f3dc9b4e5a2c2b69ce73d816d6369841318d74f15528ef6faf521291e94ae07e",
		100291: "730ceff0be2cf974e623793236fbf9ff6bac1650c67ed7f3acb88495cc4035f6",
		100292: "1f1fa80a9e3bfc64ae90c0cde68c3d5f7fbe278371b93c5005cb5c94417d51d2",
		100293: "14808b65922cf2de465921e56311bd92549a4f1c415f49b67bff28aab85c7ac1",
		100294: "ca34bff6f3082a81795283d020541bc92067a2bdead3be8486ad81f72816db50",
		100295: "4f08fc5fa41b75df637cea865951467b132af28f2678c175762c7d4c92ee30f5",
		100296: "0c93f7559df65f7bb6efaf2867bd194ce97eecf652c80e4a6d79f4652cbdbcd8",
		100297: "dbea2e488b9276d13c8c1e6f102a1c044549570075fe747a470191f510bc65eb",
		100298: "6540f390edd1ded53c70f15fad9e8ce304fc1626d77b3a8e3ca21c7a1b914061",
		100299: "4676f1289023093e346894fe68896282e99f537ccf7314cd7fb34ad919458a19",
		100300: "ceecb53ed2ab184d8a662d1c684778454088156bd473744daf2afcc1a4ddf513",
		100301: "b083744f15bec997ed70d5d460d5b34c4a08767cbe4efb4ca731161b234f2016",
		100302: "eefcfe9a4f2a3c6b8b0193b49fa0fd77777759816911ee0fa0e0f60579b5ba30",
		100303: "aeaec3f0ab5dc949ac29cf181ae65e69eac475a116d34c213daf0c0024cd91f3",
		100304: "ff14fa86d5c4fbec101ec865c739f2ea819c1238fadca91d7c82f2bd9dacdb98",
		100305: "c9067e3b40cbd1835adf5c3eafd4b639c7622de998ce36e890ef9856938f813e",
		100306: "ef78a10810e0c271e5f09c0a9ef65a2676fc3e26e25be66ccdbb0340bbac60bc",
		100307: "359183f8deaf860d53f0effd234bddcfeae640af128592fa0e81e0662925fb52",
		100308: "6b202ca8406d6184999fe2c9c68a92742e880cc7a30d29c0388f579dadc0b8dd",
		100309: "f12fcf1b8117617c8461caefa9ffec6dfa5137b70777577391d2fbad80d405e3",
		100310: "cd399f7e0e1eded8df2787de758f929a7d96be3706201c77647f52b648cbccef",
		100311: "812fee59066a47af5742002808f8d166c284a2285aadcb977c1d620d9a934e5c",
		100312: "8204610286d2c2e50a435aff315194db16557eacf1083ff4a75ebf34173ab71f",
		100313: "35409ad394ec615533702c5f5ddd8dae9fb42be3153b386820f070c7eddd7dd4",
		100314: "0e6fc02ce6136f7b23abad46a8340dabe92ceb82c1b5459caafdadac35a69514",
		100315: "03d2ca6e77984a1f2f5a0737baa185453487a7eabed2b4c9ec072f0bee70fa65",
		100316: "009d88f45286c5ded7c3285eff7a6414928ca8533f5f3ba754b5dba16b33e9aa",
		100317: "6015e69fb50b74144e3b50b293950554ad3e129964c1f766a61fe75a41968bd1",
		100318: "ed02b7f597e45b41cd6e2b6365c039d4bb9f7c311a0c60e4cabc3d2054a442b1",
		100319: "52c7bf4bab51001828a52dd7e3d12c711d2577f5ccf353f0a5c3fcba04ac1a0d",
		100320: "cd35bc214a54c088ef4282312ecf66a3e4a50d224257561aad32d443da6d019e",
		100321: "30d1e1300d08253131ff59ce32034e218367ed076f50a556e9febc820586dd1a",
		100322: "baab429625398f426bfbb3543ebd8619e7c7127015c6fd7e7f7c27b08c0105ef",
		100323: "7bd45a7ca61346d98ce99f160080717fa747d24c728650474da0b3b727bd307e",
		100324: "8731e7ad06fcf0003dafeafa26275cd09731be1299719c6d3669d363643a565b",
		100325: "81a26e53aa2bb68ae6777f769fbf6c1bd5d16ad726ea23a3738af847512cb51c",
		100326: "69f11b5e2c7e198941ef6081ab1e7178a14061265c7b05fc13e405498f5299d1",
		100327: "456b7574440bd88c9cfab3d0d57a4a3e9492fd4be5a617dd74e0b60e9b0aa585",
		100328: "72e3cf1436f8f941ea3ebc472bb590fb07169cf919f1ab792d72b2dfb42d9d93",
		100329: "a5018d3df2a4ec62032475bea831959e85ede25118fb6f18a8cebcdbcf71385c",
		100330: "5e7ba85f983a6ab0f2457c4cf145184987e1373586e8f301e7c6bd7efd0c0927",
		100331: "5d2128a93b1bf683e6dfce02200abf53b388c7be085bb013fdc3066edfeaedc0",
		100332: "b8d05441546072a5cb7cedd356ff5973040c27df0f39c33fb0fcb4f0ac4e92f5",
		100333: "953ae7368d44b28308e9c7078de0dd9316939d9dd29f0c8776184feeac66b08c",
		100334: "f0225e18676700883168b53b16414314aef90ad3b8fb831f2b127dc023cbb385",
		100335: "9229919cf57f19c80ace46d9fd77db1fb50eef84c43c1b17b73f62bcc65053be",
		100336: "b9a040dfe0e933d6487a495dc55b963f958babf1f9844aa580b2f35e5ba782a5",
		100337: "9ceb69f5e2a3989faeb7d455b05b4d6bc30ea4bad5fb6542e98c68ffaa6f5f5e",
		100338: "5725f68c290ad3e20f7638b46bf6596469d6ab4d17cdcf5b1beb94c1ef4610b9",
		100339: "6af9873235ec467277429b371af0b8c5c02c437d99039b2e0e8c59bfbedcaa78",
		100340: "c5a6047cd9717d05898cdb6808c0532881a8659516e00ab5f57a85439a39a2cc",
		100341: "4ef8e68bddf693113ad825d27d8c70034b280cdabb1725dc4bd11de046121975",
		100342: "ae1c08886240c24358422d018c79b372bd5e889b9aa4d2347053b0fb466b9849",
		100343: "0793cf94d03410b2488ea9261448b77ef0f43857c872f08a2c30b2f2470eed3d",
		100344: "6a6cae42b12009104940611047c9153dfbff7c9d49d8fd4b9996506bc7ebb05b",
		100345: "f6bdd0506641955c2811c055937b7ee31438a6a23a6a8ec05fca91834debd468",
		100346: "371ddaf41430bf5147d9ddd4f52d874b82ebcbd41ae3613a2efaa3b27189867e",
		100347: "08949650f17f75ab30f19ed0d23492e3658d0de73c0f43683b764b9792505a9f",
		100348: "b498cc2cd173ec3008fc1ebb0954e8bb83cdf20fe570990a9511bde1f2a673c0",
		100349: "1fb540777d23678e0bf68082940b9af194d0465434989b7e776b80952a2fe862",
		100350: "54ca9a73ea406e64131eb86bc826acb4b6a63be1df29b982e0298b3b642c8f66",
		100351: "442577d3b8f73895d6cebbe15e14b229acb7432c1571e6f6e2200275f1b0d912",
		100352: "c1856e0c7f63d97ff5792c65347ce6cf0672c0b9a97db1fbb93bcb9f03228c4d",
		100353: "69d9d780efd27425cd28c672a6c9522ea297aa42b0ac293ec276808c9f554772",
		100354: "240d71c900db816e220ffc01316b410f9fa5ad426f590fc582ef2ed4f24c296c",
		100355: "c9bf15d5db7abd6331a024fee199f72be0a4b04786cd96cfb3794af84558a88f",
		100356: "f96ce308359b397103e11666f5c0b102880545c0a78334c1bbcbec2ca7f36a78",
		100357: "f32a13447ff6c921cecdcc46cb935a25c85db630f5bc07f269178081362fff41",
		100358: "576118b10904d7f1d1d727bd11f76a40c68503b53f8f4f5c2d3aaf34dd8c3c28",
		100359: "ff23c017e2c3f2564fcf1ee84b92855dfe6b06f0c00bdf1e4bf4f1b1a4bb00b6",
		100360: "fec43cfc7d479ec5f49dc71e07149eeb86c2fafc7e78fdc924d9a049e24c00d1",
		100361: "6ffd6c9e74817d4a4bf233b75ac1797bdc702e25d3c914174c77b6523baab82f",
		100362: "122c7042003a0b2a90c3928d94ccf609e721017798dc4a4bc38d425170ffa252",
		100363: "e773bc7f5050fb1d4521ec571b1575f3968b4b92c7b83344675557da86f80441",
		100364: "7c74e422ca87471c916f0b3300a30fac82ba2dc825ac7a28624859fac83d1cb6",
		100365: "106709b9c301652de8baa491a55b26511a189ce4781565608adfdac0040ad73f",
		100366: "f91c29cd0b12e5478ab201ac44b00e01d0bf40a082d423ca0f154f22f5f4c3b7",
		100367: "4a12205d90d373f5052123aa2f308eebe7b5a789469fb27bf5a27d956ad9f574",
		100368: "b5bf1a8e5e410835dfd4deb7e29d51e54bc5abcf4ae2576b0ace6bee31418aea",
		100369: "5ce07c7f79fd6de30a85c1f71f3435e1874c683a9c98cf683e0d5ef0e9d9ce2b",
		100370: "b265e5134d95edfa4e6ac672382f7a391f8ae1f9ca525de02a6b0762cb081bd7",
		100371: "4cdb99292a60e1253ef6a8c2b2d011dc27cd9d353934d6f7d45bd0acc91bfb5f",
		100372: "65ddc5f4c0af7d3de5358f09235316fe5036ee8d52e44f1dd15286cfe0d521cf",
		100373: "a8ed9b4f371f314d7f7f32a12a4eaa64ae863ebf7ba1767736c424594772955b",
		100374: "0cbcc7c2dcb5325d1eb2b9f1b7f5565337db6ae261daafcbfc4f5b3d59de47ad",
		100375: "a72138dcd825b79a6a00424dca273314c894cc4aadbb4b7a953d43a7d9abd8d6",
		100376: "449f8c302d1e8d197fb9944d3d7a76fa2aa890ffb8ba51f5382171a5c252f7bd",
		100377: "fdad4c2c3f15edba2b9635202eb132aec1bd136c2e6f6190f1bc4d0da2d676d5",
		100378: "43571b3e356d15fb4418f7fba49ef324d42ed886a09189946fd2f158770307cc",
		100379: "06ab5311f58348b03e4e59d89f628cc75ddc9c7052c2c809fb48eadbe8f694d8",
		100380: "f21097b176062f74574008b2804b3f25c1b8730c7b3ed8c5fe2568c2b6caffb7",
		100381: "672bbfaf89515bc6e575aece4f6941ca112774fd3f1f85efdc559d1a8007562e",
		100382: "cf42870ff063b58a1fc633dafb77d9d28046c557677d5a7c2f296c39324d8afa",
		100383: "989a51761704f1a5262361760724458639dba4fdb994d27fd4706772d05cf8aa",
		100384: "d63ba37e24690c9fb80699e0075b6d231a9a32641ad2277b2ac6c7b69ea76b2b",
		100385: "625fcade1926c8ca007689d826d6623ad8718e7574e834c4c6cb2978abaead47",
		100386: "a4b954b81e3c660529ddad98e608451aba4e91dcbe0245e9a67794b01657f404",
		100387: "ae19b8f9fad10606a084734e9894cf6eac84da1be651f507c38f0be2fd1648fa",
		100388: "2370306792e85c61de9b3076540c35746228e2bd8ec3c547d4a1dfb9ac26a95c",
		100389: "5e946bb1921737eb4e6f225bfcd2d958cbbd71058ffb6cdd262b2130cb680f74",
		100390: "8409ae9b1fdc44bd21f2560126287ab2ac113b8caa4e0426b5822bfe81262663",
		100391: "f446c3d56aa28300991ab83c7014529f3af4d001bd60ecd01428d3cb67081d5d",
		100392: "62b85fb99d3ded0e2e31ac607097391c3c0a30e54548e31345f9e6512c847159",
		100393: "173e6b5e705abdbb9d17e94fefe8264a6bae7a9e1ccbba8bd7fb295af02c0c0a",
		100394: "f4545a36a9842ca8cc302b59b87d140549b4b75ffae473e29bb31c3e3566176c",
		100395: "4136c80318c25258f7716e444c8a8131ce422fa1166645c0ab75e01b9421c184",
		100396: "c86bdc0fc9fa2b61e2b16e70c56350650732f390a6fb91b3e11f1210d88778c4",
		100397: "b4657c0887a95daf2d71f625aefee90491b2819bfdfa10cfcefe6a6848d1f3ec",
		100398: "bc796bf69f3c29d3bb21f7496ac26ae27363958c46919e98fc0913a0d6b34368",
		100399: "6e101abf9fbcedee9dad098c4a3d8c1d8e3163a4c0ca6b1d8765890a882a4042",
		100400: "c43197e70daf24e11ca90b1f811f3a8db656f7aaa27d71289a2f7fbc5d4d5a75",
		100401: "20740d11ae540cd773e4f9e588e9e1d35f17384dca462fb00069d7b256ecc806",
		100402: "ef105bf77e27f247e958ab5309adb2762ec6e6a392ed38c2f46a7c243c4a074b",
		100403: "e55ec57bfd6cad76aacc4c6e9c4251f7a23b7cba4358b92f49fdf6b51d926440",
		100404: "796578b11caac91e55749924cc63dad7311d413ad272ebb62d4820d56cad0cb8",
		100405: "758d73e1266dbf8dbbf0b5bdab388791f8cc4f719d9c21b6959d6b947c536474",
		100406: "f04e6db920da59b540743be303aeb4a5ef03a0aca7a3b3e7726dc44dc662b4cc",
		100407: "3e390cb39126bf5d9d8a97084520f5417619d8055e3f7ac8bc65136812c5e4ea",
		100408: "b072a8d8f222c660b4d0a0a3ddcba1b2970c8f8058fb22618196646d1d254fbd",
		100409: "9a9dc042379b21cee7b321e152de386d2e13cc51eda15298884319cd6fea34eb",
		100410: "fde1d3bda8e6aaffab5bf83b701dc31c3b2238a8ef8b6fa04ebdb0e0aac7e505",
		100411: "e086ff9e553e85e3577ee1fa1bd535d6e343a07e653a09254197594946bbb11e",
		100412: "d90300531f60ae241d362cf0aeb2a99e94c5873b440242f55c8bee9a6fdb90c5",
		100413: "efc26a26188376e2ca437681ea0f4eacbdd98c0675c40383847866f589d1fa25",
		100414: "398842db091d4f781659c8741e23a78b3e19858c6a0c6dc8f5191d92ad0683c2",
		100415: "c66cdb6889031b6d7370d07ba9344df14da137dd3698f3b50130b612f360fd17",
		100416: "1dd0e8bddb23f51a198cef6a434217e8597dbcf6bbc2cce58c6055f75f4a5ecf",
		100417: "fce35dc127eb67ea360db1d2f42610c1ce569e4b6832b7c9a108c1b16004d214",
		100418: "9dd0518402b1fa8ef0528fd05af41ab264e1705dba0ec1ea8fc2bd8d07589f08",
		100419: "b98ddc690bc3b77dd6c8a491409b2d8e37320aa71d5205c7e5c581e19143fed7",
		100420: "a1f1b2b7f3554354ca8386407328d62f89510598169fd9c01f272ec31976406c",
		100421: "ba9f32b93357de925f312ff9b3a8d65bf9ef24f5854378c159eb197ed11a547e",
		100422: "aa3d7c84eb4843889ff81527806ed66dc3f5876c74fdb7f40188652f50b2e13f",
		100423: "ca35aa7255c364e1a475285334caf6660766bfad4da1b7ffdffae874e3cb1cbc",
		100424: "1e58bfac23288035f470e4900506d50ae6622dcb0aa80d8622e96d9c9fd05737",
		100425: "acddc1685784a79918894d944e884598c5f6394048c00443e08bad9f51188a38",
		100426: "5caa499d9acdad95942ce4952e08a78f5c3b3aaf308b78fda35e27730efc0217",
		100427: "33354bc8407e7e7f6b8fb8317d0aec87cb8b8c9220429adcd057052530211aa2",
		100428: "eb99b74dd5408bf09534fc5e33c822ea121fc838804c5e5a93c808058e510df3",
		100429: "93f00379b383d1278a1df1b62091fd7bb1b7bb6c1dfe4e0c722415b601165c98",
		100430: "747190eee43b88dbb9b4934d2f8b03b0fd957673d3796a25fd712ad679b56fd6",
		100431: "625fee8092baf885c6be898dfda91793545ffe70076bd114d37e6b702dc22c10",
		100432: "5fb9ae4a78f0d8088d7398e53295476d9de65f4643b2fc556a03de83df62ef28",
		100433: "5e870fa7b6ebf38c73c9a56c382d26b1d99c3eff6bdde44dedd6e37e2c84a698",
		100434: "c7dfb9dbf31c854b30ea6434209b3cc5746a5c1e248da4351a1dc3dde454f431",
		100435: "679e84f03f95c9978816559c544437a681796820de49b6ff094f739f8b51e7c4",
		100436: "80f08d99d5d3cca85a17fcf700ed844e5d60f3b1ec14bf53e139e297f2a7f19e",
		100437: "8c958fc67a2d4c2cee3672a4c4dacd9c45d654c78a4f120fd2a1e03ba654c6fa",
		100438: "c0b9d0a318374819094e81e873c1a9d5e9df1820a7161449f43c164634a53eef",
		100439: "348e215e8cb58bc6eb1a20c900a532d3e3bd3bc576233a5a36a30f5b9c6ff4e8",
		100440: "c119aafc6f1821b8dd03814a8bf9e89d3ea85c5d9449267cf0a905bfd7e22616",
		100441: "bcde81d485fed7b238d6bfe12d8ba18fdca7d1725135a6502471ff008401927a",
		100442: "d32675a059f3e75bfa356f33e5e590e71222503cf23640c2104820eee7ae8bf8",
		100443: "946f4bb207beb37ffabb0e878db49896bb4f79357f1339391ed971c8183d6431",
		100444: "c5829536820e28f8de9f0418a7a39fa7d22d207ce3a483b1e4800a485ae06179",
		100445: "0604d96049ccfb180b602c4ccada04d844e2bae04b4d407a24804460b1e37cc7",
		100446: "2da1a965f695ae88412319b2da3960c9c717cdc2d8614ef9d291d05885ad525c",
		100447: "af3ac96ebcefc8652e5456b0539d5406f634bb0abdb495bcdab85fe913556b1f",
		100448: "efa795b175186cf227219243a9e14e43844d3d9e3fe4d49d16ec5c33da21677b",
		100449: "59a57437f86bcee43c727b2a70b1159bfa01a598c417d6ffc3e8fd966f29ee8b",
		100450: "e131ad5770e1088553248ed1ff8b86e5565a03ae895dbdaab64e8955c8016ea9",
		100451: "8a8a06708f859ab693e19bb45c6462284d2da0cd0ea3aa8fb66a647855436e72",
		100452: "79bad042ae4118423db0814a85ad226a342918fd89974ed34f2e4b43c9c3cc9d",
		100453: "3941351e2e4e47c7eceab4aff5705a72313553833b5a023ce5b7d159f649a619",
		100454: "01e8297529b056c77d1799e6113cec8e78a627b6e0e47a3bc86d90678c20615c",
		100455: "db816ac9de49f54bc68c520d8cc2a11b51ba0a37587e134062a0edea1e2eee74",
		100456: "46b9316e7aaa152da0b985957d4af964cdb6bd61ecdd6312937e14b81edb88f8",
		100457: "6a5f1929f9a2bf9c75bbf2a408f235da3187bd43029e1daca2a93161799b0a93",
		100458: "857199048d34d5c0944a5dba610e3074de6a11c171deea071a2ea11f018bf4a4",
		100459: "425b1a1ce7fdb968acdac215f1978b7678dc53466b1c6aa06196fe2dbe0e93a4",
		100460: "20778f656f867ccd6d8d464aeeaf3f52e35668d01338542432afca5c2b4653ba",
		100461: "0445b760388b6bf950f53f8cebc16cd02536971ec84325450f219e4b8b328240",
		100462: "98d72d92dfd017bcaa9f6298d2f7a1d6ea60bdadb6c0026ab5de462dd6d20e7d",
		100463: "8436de7d9272530b3324d567b5946b01dcc3f98852b9bed06e01d1131d60dfd0",
		100464: "86c2af4cc7554d171342494af5ed2479db68ea7721b56bc09156178f1867de9d",
		100465: "ac73647589bc6d337279879b06387c64d2c71f6cb7c1b55fbe40bcb0806248e6",
		100466: "d4ec61c643a709aef66aaf65cfd2ff932be46c6cf311c2671de875560355649b",
		100467: "b61506914431ba22863103a2f0332c58e1eb6c7afe94fd7982af293ccfc0b9b6",
		100468: "4f4ccd9f50ff820545790021e8b6b3fcded1540ab4a43478f1bf91ed64ea0162",
		100469: "e1b11fc9ec4fe9f1d2e1d4681ac6e58cd5c2c3088d91edaea7e3efc685f0473e",
		100470: "facb5257907abbf63b58dc0a1b842b04f108b1db58d7e97f01af3cebc438ea6a",
		100471: "15dcf427cb5eeefa607cdb6007d34ff782786d7f82e7ec6c9f288360b89e049c",
		100472: "a3061685f2347bcc95be1cb991d652a5680a1527af25ac1a109fc908e07f8abb",
		100473: "eac553f32e693245ebf9a6804f80be0f31d466180b2b37256bcfea0b68be422f",
		100474: "8759852510530d1b9e8a905c4547fdd9f9bfb1cb978fcd9e897223e357223040",
		100475: "bba5ef639e567c4fa7ff61559327e60aecb0778340b9d19ed7185bb723e91e80",
		100476: "4b90ea6f2c4a07db96400480f93c36d3fc391d0ce9372454be5e124e73859b72",
		100477: "a41202ee17b964f2cd460991338efb93b5b560721aab5e62a83ff0848ab3b2d0",
		100478: "7c7f91695b2d9960dea25aa512a8dc2f0553b269c54d43d22ec3ec2c190ed405",
		100479: "96b631b4e00b65728f08b06586d577c705d6c57f609f359dfd47d7dc58b3a7ae",
		100480: "b2ad1996ec2fe72a690531d8a6711841b14076e2ed12c0b0f1c96902b0727f22",
		100481: "1bf0708a8e19f81231576fe888f3f046e80e93ad96089206c2f023761ed7c3be",
		100482: "b28531c60bc6a22dfdf359e2e777f7b740b7b1e61285f7e3fe95884ef4808a4a",
		100483: "84d80c808ab1a54a943e751b58807b201b49d4b6c3403d0a420fed88d77ca489",
		100484: "a7eeab2c731fd8a534cdf5ee0d86cfe884816e61ff263d29110890324fd4462f",
		100485: "13751f9a51bc7e2d886b508ff5e0e78fe54db62397d36cac43d56ee528b94ff5",
		100486: "1cef504196fc2e223fa18dbb870ad2da86447c36686a37bec7582841f04aca28",
		100487: "e4d896bea048c3209a8756b748e045f3eb5a01e7d18dfb16131d7358a36ae5d9",
		100488: "7dce706eb8363d20ace1fbd7a057976fa9c4f2975966b34d7aec082b2621f8a1",
		100489: "7f267696070da34836e46294b449bb8b8efe69799715108660fccf65a2ea9ec2",
		100490: "38de38345ec130af35a54a65f6152484248843b7f7df4f55187ec48f931bd9cf",
		100491: "a35a1e94d0f797dc2a19aa36bd2225d3cef5195164c5895181eb915f24d21813",
		100492: "1acedebcdd0cbc8d834a4fbf5c9b400d11cdf62baacb8cadaa581aea3a266ab1",
		100493: "f1a9388002efaacc85bc9b21d39a0ed863edac3c0fecea5d7351912fbf9cff34",
		100494: "ef298afd672b886b9fe5f9ef9c9984ab0cbaf37082042ed3d625b0d59b353735",
		100495: "d8c3af13470dfc097a618213b889eacc4065b3ccfd9e8def66af804952622259",
		100496: "12c67944a1767080b496e9dd5f04ec650ddb1f2e68eb70b62ab2e459b676cf68",
		100497: "a7c0b3a82cf27163b9e407fa141175fc47b44f73c5e0d895a2728aef13e1d0f5",
		100498: "296a196bd4ac50192b5425d2e7a3ab07510f9282cf74532b9909c22718d3d9e5",
		100499: "87d0fde9474669fb3ab8e4f288b92901b3847ae2280b07610d3d0d0442d5088c",
	}
)

func initContinuousBlockHashes() map[uint32]*common.Uint256 {
	if !hasWorkbenchOnlyFlag() {
		return nil
	}
	hashes := make(map[uint32]*common.Uint256)
	for k, v := range continuousBlockHashesRaw {
		hash, _ := common.Uint256FromHexString(v)
		hashes[k] = hash
	}
	return hashes
}

func initRandomBlockHashes() map[uint32]*common.Uint256 {
	if !hasWorkbenchOnlyFlag() {
		return nil
	}
	hashes := make(map[uint32]*common.Uint256)
	for k, v := range randomBlockHashesRaw {
		hash, _ := common.Uint256FromHexString(v)
		hashes[k] = hash
	}
	return hashes
}

func initStopHash() *common.Uint256 {
	if !hasWorkbenchOnlyFlag() {
		return nil
	}
	hash, _ := common.Uint256FromHexString(
		"a78f6af2e9fb078c3f762813cd8461533f8c70cd9233d4a27e965b299f476268")
	return hash
}

func hasWorkbenchOnlyFlag() bool {
	if flagParsed {
		return isBenchOnly
	}
	flag.Parse()
	isBenchOnly = *benchOnlyFlag != ""
	return isBenchOnly
}
