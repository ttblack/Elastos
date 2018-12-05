package api

import (
	"github.com/elastos/Elastos.ELA/blockchain"
	"github.com/elastos/Elastos.ELA/core/types"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/yuin/gopher-lua"
)

const (
	luaArbitratorsTypeName = "arbitrators"
	MajorityCount          = 3
)

var (
	arbitratorsPublicKeys = []string{
		"023a133480176214f88848c6eaa684a54b316849df2b8570b57f3a917f19bbc77a",
		"030a26f8b4ab0ea219eb461d1e454ce5f0bd0d289a6a64ffc0743dab7bd5be0be9",
		"0288e79636e41edce04d4fa95d8f62fed73a76164f8631ccc42f5425f960e4a0c7",
		"03e281f89d85b3a7de177c240c4961cb5b1f2106f09daa42d15874a38bbeae85dd",
		"0393e823c2087ed30871cbea9fa5121fa932550821e9f3b17acef0e581971efab0",
	}
	arbitratorsPrivateKeys = []string{
		"e372ca1032257bb4be1ac99c4861ec542fd55c25c37f5f58ba8b177850b3fdeb",
		"e6deed7e23406e2dce7b01e85bcb33872a47b6200ca983fcf0540dff284923b0",
		"4441968d02a5df4dbc08ca11da2acc86c980e5fe9ff250450a80fd7421d2b0f1",
		"0b14a04e203301809feccc61dbf4e745203a3263d29a4b4091aaa138ba5fb26d",
		"0c11ebca60af2a09ac13dd84fd29c03b99cd086a08a69a9e5b87255fd9cf2eee",
	}
)

func RegisterArbitratorsType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaArbitratorsTypeName)
	L.SetGlobal("arbitrators", mt)
	// static attributes
	L.SetField(mt, "new", L.NewFunction(newArbitrators))
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), arbitratorsMethods))
}

// Constructor
func newArbitrators(L *lua.LState) int {
	arbitersByte := make([][]byte, 0)
	for _, arbiter := range arbitratorsPublicKeys {
		arbiterByte, _ := common.HexStringToBytes(arbiter)
		arbitersByte = append(arbitersByte, arbiterByte)
	}

	a := &arbitrators{arbitersByte, 0}

	ud := L.NewUserData()
	ud.Value = a
	L.SetMetatable(ud, L.GetTypeMetatable(luaArbitratorsTypeName))
	L.Push(ud)

	return 1
}

// Checks whether the first lua argument is a *LUserData with *Attribute and returns this *Attribute.
func checkArbitrators(L *lua.LState, idx int) *arbitrators {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*arbitrators); ok {
		return v
	}
	L.ArgError(1, "arbitrators expected")
	return nil
}

var arbitratorsMethods = map[string]lua.LGFunction{
	"get_duty_index": arbitratorsGetDutyIndex,
	"set_duty_index": arbitratorsSetDutyIndex,
}

func arbitratorsGetDutyIndex(L *lua.LState) int {
	a := checkArbitrators(L, 1)
	L.Push(lua.LNumber(a.dutyChangedCount))

	return 1
}

func arbitratorsSetDutyIndex(L *lua.LState) int {
	a := checkArbitrators(L, 1)
	a.dutyChangedCount = uint32(L.ToInt(2))

	return 0
}

//mock object of arbitrators
type arbitrators struct {
	currentArbitrators [][]byte
	dutyChangedCount   uint32
}

func (a *arbitrators) OnBlockReceived(b *types.Block, confirmed bool) {
	panic("implement me")
}

func (a *arbitrators) OnConfirmReceived(p *types.DPosProposalVoteSlot) {
	panic("implement me")
}

func (a *arbitrators) StartUp() error {
	panic("implement me")
}

func (a *arbitrators) ForceChange() error {
	panic("implement me")
}

func (a *arbitrators) GetArbitrators() [][]byte {
	return a.currentArbitrators
}

func (a *arbitrators) GetCandidates() [][]byte {
	panic("implement me")
}

func (a *arbitrators) GetNextArbitrators() [][]byte {
	panic("implement me")
}

func (a *arbitrators) GetNextCandidates() [][]byte {
	panic("implement me")
}

func (a *arbitrators) GetArbitratorsProgramHashes() []*common.Uint168 {
	panic("implement me")
}

func (a *arbitrators) GetCandidatesProgramHashes() []*common.Uint168 {
	panic("implement me")
}

func (a *arbitrators) GetOnDutyArbitrator() []byte {
	return a.GetNextOnDutyArbitrator(0)
}

func (a *arbitrators) GetNextOnDutyArbitrator(offset uint32) []byte {
	index := (a.dutyChangedCount + offset) % uint32(len(a.currentArbitrators))
	return a.currentArbitrators[index]
}

func (a *arbitrators) HasArbitersMajorityCount(num uint32) bool {
	//note "num > MajorityCount" in real logic
	return num >= MajorityCount
}

func (a *arbitrators) HasArbitersMinorityCount(num uint32) bool {
	//note "num >= uint32(len(arbitratorsPublicKeys))-MajorityCount" in real logic
	return num > uint32(len(arbitratorsPublicKeys))-MajorityCount
}

func (a *arbitrators) RegisterListener(listener blockchain.ArbitratorsListener) {
	panic("implement me")
}

func (a *arbitrators) UnregisterListener(listener blockchain.ArbitratorsListener) {
	panic("implement me")
}
