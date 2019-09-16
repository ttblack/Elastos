// Copyright (c) 2012-2018 The Elastos Open Source Project
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

#include "MainchainSubWallet.h"
#include "MasterWallet.h"

#include <SDK/Common/Utils.h>
#include <SDK/Common/ErrorChecker.h>
#include <SDK/WalletCore/KeyStore/CoinInfo.h>
#include <SDK/Wallet/UTXO.h>
#include <SDK/Plugin/Transaction/Asset.h>
#include <SDK/Plugin/Transaction/Payload/TransferCrossChainAsset.h>
#include <SDK/Plugin/Transaction/Payload/ProducerInfo.h>
#include <SDK/Plugin/Transaction/Payload/CancelProducer.h>
#include <SDK/Plugin/Transaction/Payload/OutputPayload/PayloadVote.h>
#include <SDK/Plugin/Transaction/Payload/CRInfo.h>
#include <SDK/Plugin/Transaction/Payload/UnregisterCR.h>
#include <SDK/Plugin/Transaction/TransactionInput.h>
#include <SDK/Plugin/Transaction/TransactionOutput.h>
#include <SDK/SpvService/Config.h>
#include <CMakeConfig.h>

#include <vector>
#include <map>
#include <boost/scoped_ptr.hpp>

namespace Elastos {
	namespace ElaWallet {

#define DEPOSIT_MIN_ELA 5000

		MainchainSubWallet::MainchainSubWallet(const CoinInfoPtr &info,
											   const ChainConfigPtr &config,
											   MasterWallet *parent) :
				SubWallet(info, config, parent) {
		}

		MainchainSubWallet::~MainchainSubWallet() {

		}

		nlohmann::json MainchainSubWallet::CreateDepositTransaction(const std::string &fromAddress,
																	const std::string &lockedAddress,
																	const std::string &amount,
																	const std::string &sideChainAddress,
																	const std::string &memo) {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("lockedAddr: {}", lockedAddress);
			ArgInfo("amount: {}", amount);
			ArgInfo("sideChainAddr: {}", sideChainAddress);
			ArgInfo("memo: {}", memo);
			BigInt value;
			value.setDec(amount);


			PayloadPtr payload = nullptr;
			try {
				TransferInfo info(sideChainAddress, 0, value);
				payload = PayloadPtr(new TransferCrossChainAsset({info}));
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
												  "Side chain message error: " + std::string(e.what()));
			}

			std::vector<OutputPtr> outputs;
			Address receiveAddr(lockedAddress);
			outputs.emplace_back(OutputPtr(new TransactionOutput(value + _config->MinFee(), receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::transferCrossChainAsset, payload);

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		TransactionPtr MainchainSubWallet::CreateVoteTx(const VoteContent &voteContent, const std::string &memo, bool max) {
			std::string m;

			if (!memo.empty())
				m = "type:text,msg:" + memo;

			TransactionPtr tx = _walletManager->GetWallet()->Vote(voteContent, m, max);

			if (_info->GetChainID() == "ELA")
				tx->SetVersion(Transaction::TxVersion::V09);

			tx->FixIndex();

			return tx;
		}

		nlohmann::json MainchainSubWallet::GenerateProducerPayload(
			const std::string &ownerPublicKey,
			const std::string &nodePublicKey,
			const std::string &nickName,
			const std::string &url,
			const std::string &ipAddress,
			uint64_t location,
			const std::string &payPasswd) const {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("ownerPubKey: {}", ownerPublicKey);
			ArgInfo("nodePubKey: {}", nodePublicKey);
			ArgInfo("nickName: {}", nickName);
			ArgInfo("url: {}", url);
			ArgInfo("ipAddress: {}", ipAddress);
			ArgInfo("location: {}", location);
			ArgInfo("payPasswd: *");

			ErrorChecker::CheckPassword(payPasswd, "Generate payload");

			Key verifyPubKey;
			bytes_t ownerPubKey = bytes_t(ownerPublicKey);
			verifyPubKey.SetPubKey(ownerPubKey);

			bytes_t nodePubKey = bytes_t(nodePublicKey);
			verifyPubKey.SetPubKey(nodePubKey);

			ProducerInfo pr;
			pr.SetPublicKey(ownerPubKey);
			pr.SetNodePublicKey(nodePubKey);
			pr.SetNickName(nickName);
			pr.SetUrl(url);
			pr.SetAddress(ipAddress);
			pr.SetLocation(location);

			ByteStream ostream;
			pr.SerializeUnsigned(ostream, 0);
			bytes_t prUnsigned = ostream.GetBytes();

			Key key = _subAccount->DeriveOwnerKey(payPasswd);
			pr.SetSignature(key.Sign(prUnsigned));

			nlohmann::json payloadJson = pr.ToJson(0);

			ArgInfo("r => {}", payloadJson.dump());
			return payloadJson;
		}

		nlohmann::json MainchainSubWallet::GenerateCancelProducerPayload(
			const std::string &ownerPublicKey,
			const std::string &payPasswd) const {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("ownerPubKey: {}", ownerPublicKey);
			ArgInfo("payPasswd: *");

			ErrorChecker::CheckPassword(payPasswd, "Generate payload");
			size_t pubKeyLen = ownerPublicKey.size() >> 1;
			ErrorChecker::CheckParam(pubKeyLen != 33 && pubKeyLen != 65, Error::PubKeyLength,
									 "Public key length should be 33 or 65 bytes");

			CancelProducer pc;
			pc.SetPublicKey(ownerPublicKey);

			ByteStream ostream;
			pc.SerializeUnsigned(ostream, 0);
			bytes_t pcUnsigned = ostream.GetBytes();

			Key key = _subAccount->DeriveOwnerKey(payPasswd);
			pc.SetSignature(key.Sign(pcUnsigned));

			nlohmann::json payloadJson = pc.ToJson(0);
			ArgInfo("r => {}", payloadJson.dump());
			return payloadJson;
		}

		nlohmann::json MainchainSubWallet::CreateRegisterProducerTransaction(
			const std::string &fromAddress,
			const nlohmann::json &payloadJson,
			const std::string &amount,
			const std::string &memo) {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("payload: {}", payloadJson.dump());
			ArgInfo("amount: {}", amount);
			ArgInfo("memo: {}", memo);

			BigInt bgAmount, minAmount(DEPOSIT_MIN_ELA);
			bgAmount.setDec(amount);

			minAmount *= SELA_PER_ELA;

			ErrorChecker::CheckParam(bgAmount < minAmount, Error::DepositAmountInsufficient,
									 "Producer deposit amount is insufficient");

			PayloadPtr payload = PayloadPtr(new ProducerInfo());
			try {
				payload->FromJson(payloadJson, 0);
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
												  "Payload format err: " + std::string(e.what()));
			}

			bytes_t pubkey = static_cast<ProducerInfo *>(payload.get())->GetPublicKey();
			std::string toAddress = Address(PrefixDeposit, pubkey).String();

			std::vector<OutputPtr> outputs;
			Address receiveAddr(toAddress);
			outputs.push_back(OutputPtr(new TransactionOutput(bgAmount, receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::registerProducer, payload);

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		nlohmann::json MainchainSubWallet::CreateUpdateProducerTransaction(
			const std::string &fromAddress,
			const nlohmann::json &payloadJson,
			const std::string &memo) {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("payload: {}", payloadJson.dump());
			ArgInfo("memo: {}", memo);

			PayloadPtr payload = PayloadPtr(new ProducerInfo());
			try {
				payload->FromJson(payloadJson, 0);
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
												  "Payload format err: " + std::string(e.what()));
			}

			std::vector<OutputPtr> outputs;
			Address receiveAddr(CreateAddress());
			outputs.push_back(OutputPtr(new TransactionOutput(BigInt(0), receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::updateProducer, payload);

			if (tx->GetOutputs().size() > 1) {
				tx->RemoveOutput(tx->GetOutputs().front());
				tx->FixIndex();
			}

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		nlohmann::json MainchainSubWallet::CreateCancelProducerTransaction(
			const std::string &fromAddress,
			const nlohmann::json &payloadJson,
			const std::string &memo) {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("payload: {}", payloadJson.dump());
			ArgInfo("memo: {}", memo);

			PayloadPtr payload = PayloadPtr(new CancelProducer());
			try {
				payload->FromJson(payloadJson, 0);
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
												  "Payload format err: " + std::string(e.what()));
			}

			std::vector<OutputPtr> outputs;
			Address receiveAddr(CreateAddress());
			outputs.push_back(OutputPtr(new TransactionOutput(BigInt(0), receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::cancelProducer, payload);

			if (tx->GetOutputs().size() > 1) {
				tx->RemoveOutput(tx->GetOutputs().front());
				tx->FixIndex();
			}

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		nlohmann::json MainchainSubWallet::CreateRetrieveDepositTransaction(
			const std::string &amount,
			const std::string &memo) {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("amount: {}", amount);
			ArgInfo("memo: {}", memo);

			BigInt bgAmount;
			bgAmount.setDec(amount);

			std::string fromAddress = _walletManager->GetWallet()->GetOwnerDepositAddress().String();

			std::vector<OutputPtr> outputs;
			Address receiveAddr(CreateAddress());
			outputs.push_back(OutputPtr(new TransactionOutput(bgAmount, receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::returnDepositCoin);

			if (tx->GetOutputs().size() > 1) {
				tx->RemoveOutput(tx->GetOutputs().back());
				tx->FixIndex();
			}

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		std::string MainchainSubWallet::GetOwnerPublicKey() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			std::string publicKey = _walletManager->GetWallet()->GetOwnerPublilcKey().getHex();
			ArgInfo("r => {}", publicKey);
			return publicKey;
		}

		std::string MainchainSubWallet::GetOwnerAddress() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());

			std::string address = _walletManager->GetWallet()->GetOwnerAddress().String();

			ArgInfo("r => {}", address);

			return address;
		}

		nlohmann::json MainchainSubWallet::CreateVoteProducerTransaction(
			const std::string &fromAddress,
			const std::string &stake,
			const nlohmann::json &publicKeys,
			const std::string &memo) {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("stake: {}", stake);
			ArgInfo("pubkeys: {}", publicKeys.dump());
			ArgInfo("memo: {}", memo);

			bool max = false;
			BigInt bgStake;
			if (stake == "-1") {
				max = true;
				bgStake = 0;
			} else {
				bgStake.setDec(stake);
			}

			ErrorChecker::CheckJsonArray(publicKeys, 1, "Candidates public keys");
			// -1 means max
			ErrorChecker::CheckParam(bgStake <= 0, Error::Code::VoteStakeError, "Vote stake should not be zero");

			VoteContent voteContent(VoteContent::Delegate);
			for (nlohmann::json::const_iterator it = publicKeys.cbegin(); it != publicKeys.cend(); ++it) {
				if (!(*it).is_string()) {
					ErrorChecker::ThrowParamException(Error::Code::JsonFormatError,
													  "Vote produce public keys is not string");
				}
				// Check public key is valid later
				voteContent.AddCandidate(CandidateVotes((*it).get<std::string>(), bgStake.getUint64()));
			}

			ErrorChecker::CheckParam(voteContent.GetCandidateVotes().empty(), Error::InvalidArgument,
									 "Candidate vote list should not be empty");

			TransactionPtr tx = CreateVoteTx(voteContent, memo, max);

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		nlohmann::json MainchainSubWallet::GetVotedProducerList() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());

			WalletPtr wallet = _walletManager->GetWallet();
			UTXOArray utxos = wallet->GetVoteUTXO();
			nlohmann::json j;
			std::map<std::string, uint64_t> votedList;

			for (size_t i = 0; i < utxos.size(); ++i) {
				const OutputPtr &output = utxos[i]->Output();
				if (output->GetType() != TransactionOutput::VoteOutput) {
					continue;
				}

				const PayloadVote *pv = dynamic_cast<const PayloadVote *>(output->GetPayload().get());
				if (pv == nullptr) {
					continue;
				}

				uint64_t stake = output->Amount().getUint64();
				uint8_t version = pv->Version();
				const std::vector<VoteContent> &voteContents = pv->GetVoteContent();
				std::for_each(voteContents.cbegin(), voteContents.cend(),
							  [&votedList, &stake, &version](const VoteContent &vc) {
								  if (vc.GetType() == VoteContent::Type::Delegate) {
									  std::for_each(vc.GetCandidateVotes().cbegin(), vc.GetCandidateVotes().cend(),
													[&votedList, &stake, &version](const CandidateVotes &cvs) {
														std::string c = cvs.GetCandidate().getHex();
														uint64_t votes;

														if (version == VOTE_PRODUCER_CR_VERSION)
															votes = cvs.GetVotes();
														else
															votes = stake;

														if (votedList.find(c) != votedList.end()) {
															votedList[c] += votes;
														} else {
															votedList[c] = votes;
														}
													});
								  }
							  });

			}

			j = votedList;

			ArgInfo("r => {}", j.dump());

			return j;
		}

		nlohmann::json MainchainSubWallet::GetRegisteredProducerInfo() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());

			std::vector<TransactionPtr> allTxs = _walletManager->GetWallet()->GetAllTransactions();
			nlohmann::json j;

			j["Status"] = "Unregistered";
			j["Info"] = nlohmann::json();
			for (size_t i = 0; i < allTxs.size(); ++i) {
				if (allTxs[i]->GetBlockHeight() == TX_UNCONFIRMED) {
					continue;
				}

				if (allTxs[i]->GetTransactionType() == Transaction::registerProducer ||
				    allTxs[i]->GetTransactionType() == Transaction::updateProducer) {
					const ProducerInfo *pinfo = dynamic_cast<const ProducerInfo *>(allTxs[i]->GetPayload());
					if (pinfo) {
						nlohmann::json info;

						info["OwnerPublicKey"] = pinfo->GetPublicKey().getHex();
						info["NodePublicKey"] = pinfo->GetNodePublicKey().getHex();
						info["NickName"] = pinfo->GetNickName();
						info["URL"] = pinfo->GetUrl();
						info["Location"] = pinfo->GetLocation();
						info["Address"] = pinfo->GetAddress();

						j["Status"] = "Registered";
						j["Info"] = info;
					}
				} else if (allTxs[i]->GetTransactionType() == Transaction::cancelProducer) {
					const CancelProducer *pc = dynamic_cast<const CancelProducer *>(allTxs[i]->GetPayload());
					if (pc) {
						uint32_t lastBlockHeight = _walletManager->GetWallet()->LastBlockHeight();

						nlohmann::json info;

						info["Confirms"] = allTxs[i]->GetConfirms(lastBlockHeight);

						j["Status"] = "Canceled";
						j["Info"] = info;
					}
				} else if (allTxs[i]->GetTransactionType() == Transaction::returnDepositCoin) {
					j["Status"] = "ReturnDeposit";
					j["Info"] = nlohmann::json();
				}
			}

			ArgInfo("r => {}", j.dump());
			return j;
		}

		std::string MainchainSubWallet::GetCROwnerDID() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			bytes_t pubKey = _subAccount->DIDPubKey();
			std::string addr = Address(PrefixIDChain, pubKey).String();

			ArgInfo("r => {}", addr);
			return addr;
		}

		std::string MainchainSubWallet::GetCROwnerPublicKey() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			std::string pubkey = _subAccount->DIDPubKey().getHex();
			ArgInfo("r => {}", pubkey);
			return pubkey;
		}

		nlohmann::json MainchainSubWallet::GenerateCRInfoPayload(
				const std::string &crPublicKey,
				const std::string &nickName,
				const std::string &url,
				uint64_t location,
				const std::string &payPasswd) const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("crPublicKey: {}", crPublicKey);
			ArgInfo("nickName: {}", nickName);
			ArgInfo("url: {}", url);
			ArgInfo("location: {}", location);
			ArgInfo("payPasswd: *");

			ErrorChecker::CheckPassword(payPasswd, "Generate payload");
			size_t pubKeyLen = crPublicKey.size() >> 1;
			ErrorChecker::CheckParam(pubKeyLen != 33 && pubKeyLen != 65, Error::PubKeyLength,
			                         "Public key length should be 33 or 65 bytes");

			bytes_t pubkey(crPublicKey);

			Address address(PrefixStandard, pubkey);

			CRInfo crInfo;
			crInfo.SetCode(address.RedeemScript());
			crInfo.SetNickName(nickName);
			crInfo.SetUrl(url);
			crInfo.SetLocation(location);

			Address did;
			did.SetRedeemScript(PrefixIDChain, crInfo.GetCode());
			crInfo.SetDID(did.ProgramHash());

			ByteStream ostream;
			crInfo.SerializeUnsigned(ostream, 0);
			bytes_t prUnsigned = ostream.GetBytes();

			Key key = _subAccount->DeriveDIDKey(payPasswd);

			crInfo.SetSignature(key.Sign(prUnsigned));

			nlohmann::json payloadJson = crInfo.ToJson(0);

			ArgInfo("r => {}", payloadJson.dump());
			return payloadJson;
		}

		nlohmann::json MainchainSubWallet::GenerateUnregisterCRPayload(
				const std::string &crPublicKey,
				const std::string &payPasswd) const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("publicKey: {}", crPublicKey);
			ArgInfo("payPasswd: *");

			ErrorChecker::CheckPassword(payPasswd, "Generate payload");
			size_t pubKeyLen = crPublicKey.size() >> 1;
			ErrorChecker::CheckParam(pubKeyLen != 33 && pubKeyLen != 65, Error::PubKeyLength,
			                         "Public key length should be 33 or 65 bytes");

			Key pubKey;
			pubKey.SetPubKey(bytes_t(crPublicKey));

			Address address(PrefixStandard, pubKey.PubKey());

			UnregisterCR unregisterCR;
			unregisterCR.SetCode(address.RedeemScript());

			ByteStream ostream;
			unregisterCR.SerializeUnsigned(ostream, 0);
			bytes_t prUnsigned = ostream.GetBytes();

			Key key = _subAccount->DeriveDIDKey(payPasswd);

			unregisterCR.SetSignature(key.Sign(prUnsigned));

			nlohmann::json payloadJson = unregisterCR.ToJson(0);

			ArgInfo("r => {}", payloadJson.dump());
			return payloadJson;
		}

		nlohmann::json MainchainSubWallet::CreateRegisterCRTransaction(
				const std::string &fromAddress,
				const nlohmann::json &payload,
				const std::string &amount,
				const std::string &memo) {

			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("payload: {}", payload.dump());
			ArgInfo("amount: {}", amount);
			ArgInfo("memo: {}", memo);

			BigInt bgAmount, minAmount(DEPOSIT_MIN_ELA);
			bgAmount.setDec(amount);

			minAmount *= SELA_PER_ELA;

			ErrorChecker::CheckParam(bgAmount < minAmount, Error::DepositAmountInsufficient,
			                         "cr deposit amount is insufficient");

			PayloadPtr payloadPtr = PayloadPtr(new CRInfo());
			try {
				payloadPtr->FromJson(payload, 0);
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
				                                  "Payload format err: " + std::string(e.what()));
			}

			bytes_t code = static_cast<CRInfo *>(payloadPtr.get())->GetCode();
			Address receiveAddr;
			receiveAddr.SetRedeemScript(PrefixDeposit, code);

			std::vector<OutputPtr> outputs;
			outputs.push_back(OutputPtr(new TransactionOutput(bgAmount, receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::registerCR, payloadPtr);

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		nlohmann::json MainchainSubWallet::CreateUpdateCRTransaction(
				const std::string &fromAddress,
				const nlohmann::json &payload,
				const std::string &memo) {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("payload: {}", payload.dump());
			ArgInfo("memo: {}", memo);

			PayloadPtr payloadPtr = PayloadPtr(new CRInfo());
			try {
				payloadPtr->FromJson(payload, 0);
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
				                                  "Payload format err: " + std::string(e.what()));
			}

			std::vector<OutputPtr> outputs;
			Address receiveAddr(CreateAddress());
			outputs.push_back(OutputPtr(new TransactionOutput(BigInt(0), receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::updateCR, payloadPtr);

			if (tx->GetOutputs().size() > 1) {
				tx->RemoveOutput(tx->GetOutputs().front());
				tx->FixIndex();
			}

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;

		}

		nlohmann::json MainchainSubWallet::CreateUnregisterCRTransaction(
				const std::string &fromAddress,
				const nlohmann::json &payload,
				const std::string &memo) {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("payload: {}", payload.dump());
			ArgInfo("memo: {}", memo);

			PayloadPtr payloadPtr = PayloadPtr(new UnregisterCR());
			try {
				payloadPtr->FromJson(payload, 0);
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
				                                  "Payload format err: " + std::string(e.what()));
			}

			std::vector<OutputPtr> outputs;
			Address receiveAddr(CreateAddress());
			outputs.push_back(OutputPtr(new TransactionOutput(BigInt(0), receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::unregisterCR, payloadPtr);

			if (tx->GetOutputs().size() > 1) {
				tx->RemoveOutput(tx->GetOutputs().front());
				tx->FixIndex();
			}

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());
			return result;
		}

		nlohmann::json MainchainSubWallet::CreateRetrieveCRDepositTransaction(
				const std::string &amount,
				const std::string &memo) {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("amount: {}", amount);
			ArgInfo("memo: {}", memo);

			BigInt bgAmount;
			bgAmount.setDec(amount);

			Address fromAddress = Address(PrefixDeposit, _subAccount->DIDPubKey());

			std::vector<OutputPtr> outputs;
			Address receiveAddr(CreateAddress());
			outputs.push_back(OutputPtr(new TransactionOutput(bgAmount, receiveAddr)));

			TransactionPtr tx = CreateTx(fromAddress.String(), outputs, memo);

			tx->SetTransactionType(Transaction::returnCRDepositCoin);

			if (tx->GetOutputs().size() > 1) {
				tx->RemoveOutput(tx->GetOutputs().back());
				tx->FixIndex();
			}

			nlohmann::json result;
			EncodeTx(result, tx);
			ArgInfo("r => {}", result.dump());
			return result;
		}

		nlohmann::json MainchainSubWallet::CreateVoteCRTransaction(
				const std::string &fromAddress,
				const nlohmann::json &votes,
				const std::string &memo) {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());
			ArgInfo("fromAddr: {}", fromAddress);
			ArgInfo("votes: {}", votes.dump());
			ArgInfo("memo: {}", memo);

			ErrorChecker::CheckParam(!votes.is_object(), Error::Code::JsonFormatError, "votes is error json format");

			BigInt bgStake = 0;

			VoteContent voteContent(VoteContent::CRC);
			std::vector<CandidateVotes> candidates;
			for (nlohmann::json::const_iterator it = votes.cbegin(); it != votes.cend(); ++it) {
				std::string pubkey = it.key();
				uint64_t value = it.value().get<std::uint64_t>();

				Address address(PrefixStandard, pubkey);
				voteContent.AddCandidate(CandidateVotes(address.RedeemScript(), value));
			}

			TransactionPtr tx = CreateVoteTx(voteContent, memo, false);

			nlohmann::json result;
			EncodeTx(result, tx);

			ArgInfo("r => {}", result.dump());

			return result;
		}

		nlohmann::json MainchainSubWallet::GetVotedCRList() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());

			WalletPtr wallet = _walletManager->GetWallet();
			UTXOArray utxos = wallet->GetVoteUTXO();
			nlohmann::json j;
			std::map<std::string, uint64_t> votedList;

			for (size_t i = 0; i < utxos.size(); ++i) {
				const OutputPtr &output = utxos[i]->Output();
				if (output->GetType() != TransactionOutput::VoteOutput) {
					continue;
				}

				const PayloadVote *pv = dynamic_cast<const PayloadVote *>(output->GetPayload().get());
				if (pv == nullptr) {
					continue;
				}

				uint64_t stake = output->Amount().getUint64();
				const std::vector<VoteContent> &voteContents = pv->GetVoteContent();
				std::for_each(voteContents.cbegin(), voteContents.cend(),
				              [&votedList, &stake](const VoteContent &vc) {
					              if (vc.GetType() == VoteContent::Type::CRC) {
						              std::for_each(vc.GetCandidateVotes().cbegin(), vc.GetCandidateVotes().cend(),
						                            [&votedList, &stake](const CandidateVotes &candidate) {
							                            std::string c = candidate.GetCandidate().getHex();
							                            if (votedList.find(c) != votedList.end()) {
								                            votedList[c] += candidate.GetVotes();
							                            } else {
								                            votedList[c] = candidate.GetVotes();
							                            }
						                            });
					              }
				              });

			}

			j = votedList;

			ArgInfo("r => {}", j.dump());

			return j;
		}

		nlohmann::json MainchainSubWallet::GetRegisteredCRInfo() const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());

			std::vector<TransactionPtr> allTxs = _walletManager->GetWallet()->GetAllTransactions();
			nlohmann::json j;

			j["Status"] = "Unregistered";
			j["Info"] = nlohmann::json();
			for (size_t i = 0; i < allTxs.size(); ++i) {
				if (allTxs[i]->GetBlockHeight() == TX_UNCONFIRMED) {
					continue;
				}

				if (allTxs[i]->GetTransactionType() == Transaction::registerCR ||
					allTxs[i]->GetTransactionType() == Transaction::updateCR) {
					const CRInfo *pinfo = dynamic_cast<const CRInfo *>(allTxs[i]->GetPayload());
					if (pinfo) {
						nlohmann::json info;
						ByteStream stream(pinfo->GetCode());
						bytes_t pubKey;
						stream.ReadVarBytes(pubKey);

						info["CROwnerPublicKey"] = pubKey.getHex();
						info["CROwnerDID"] = pinfo->GetDID().GetHex();
						info["NickName"] = pinfo->GetNickName();
						info["Url"] = pinfo->GetUrl();
						info["Location"] = pinfo->GetLocation();

						j["Status"] = "Registered";
						j["Info"] = info;
					}
				} else if (allTxs[i]->GetTransactionType() == Transaction::unregisterCR) {
					const UnregisterCR *pc = dynamic_cast<const UnregisterCR *>(allTxs[i]->GetPayload());
					if (pc) {
						uint32_t lastBlockHeight = _walletManager->GetWallet()->LastBlockHeight();

						nlohmann::json info;

						info["Confirms"] = allTxs[i]->GetConfirms(lastBlockHeight);

						j["Status"] = "Canceled";
						j["Info"] = info;
					}
				} else if (allTxs[i]->GetTransactionType() == Transaction::returnCRDepositCoin) {
					j["Status"] = "ReturnDeposit";
					j["Info"] = nlohmann::json();
				}
			}

			ArgInfo("r => {}", j.dump());
			return j;
		}

		nlohmann::json MainchainSubWallet::GetVoteInfo(const std::string &type) const {
			ArgInfo("{} {}", _walletManager->GetWallet()->GetWalletID(), GetFunName());

			WalletPtr wallet = _walletManager->GetWallet();
			UTXOArray utxos = wallet->GetVoteUTXO();
			nlohmann::json jinfo;
			time_t timestamp;

			std::map<std::string, uint64_t> votedList;

			for (UTXOArray::iterator u = utxos.begin(); u != utxos.end(); ++u) {
				const OutputPtr &output = (*u)->Output();
				if (output->GetType() != TransactionOutput::VoteOutput) {
					continue;
				}

				TransactionPtr tx = wallet->TransactionForHash((*u)->Hash());
				assert(tx != nullptr);
				timestamp = tx->GetTimestamp();

				const PayloadVote *pv = dynamic_cast<const PayloadVote *>(output->GetPayload().get());
				if (pv == nullptr) {
					continue;
				}

				const std::vector<VoteContent> &voteContents = pv->GetVoteContent();
				std::for_each(voteContents.cbegin(), voteContents.cend(),
							  [&jinfo, &type, &timestamp](const VoteContent &vc) {
								  nlohmann::json j;
								  if (type.empty() || type == vc.GetTypeString()) {
									  if (vc.GetType() == VoteContent::CRC)
										  j["Amount"] = vc.GetTotalVoteAmount();
									  else if (vc.GetType() == VoteContent::Delegate)
										  j["Amount"] = vc.GetMaxVoteAmount();
									  j["Type"] = vc.GetTypeString();
									  j["Timestamp"] = timestamp;
									  j["Expiry"] = nlohmann::json();
									  if (!type.empty()) {
										  nlohmann::json candidateVotes;
										  std::for_each(vc.GetCandidateVotes().cbegin(), vc.GetCandidateVotes().cend(),
														[&candidateVotes](const CandidateVotes &cv) {
															std::string c = cv.GetCandidate().getHex();
															candidateVotes[c] = cv.GetVotes();
														});
										  j["Votes"] = candidateVotes;
									  }
									  jinfo.push_back(j);
								  }
							  });

			}

			ArgInfo("r => {}", jinfo.dump());

			return jinfo;
		}

	}
}
