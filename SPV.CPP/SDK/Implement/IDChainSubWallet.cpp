// Copyright (c) 2012-2018 The Elastos Open Source Project
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

#include "IDChainSubWallet.h"
#include "MasterWallet.h"

#include <SDK/Common/ErrorChecker.h>
#include <SDK/Common/Utils.h>
#include <SDK/Common/Log.h>
#include <SDK/WalletCore/KeyStore/CoinInfo.h>
#include <SDK/Plugin/Transaction/Payload/PayloadRegisterIdentification.h>

#include <set>
#include <boost/scoped_ptr.hpp>

#define ID_REGISTER_BUFFER_COUNT 100

namespace Elastos {
	namespace ElaWallet {

		IDChainSubWallet::IDChainSubWallet(const CoinInfoPtr &info,
										   const ChainConfigPtr &config,
										   MasterWallet *parent) :
				SidechainSubWallet(info, config, parent) {

			std::vector<std::string> registeredIds = _parent->GetAllIDs();
			if (registeredIds.size() != 1) {
				if (_subAccount->Parent()->GetSignType() == Account::Standard) {
					registeredIds.clear();
					registeredIds.push_back(_parent->DeriveIDAndKeyForPurpose(0, 0));
				}
			}

			_walletManager->getWallet()->InitListeningAddresses(registeredIds);
		}

		IDChainSubWallet::~IDChainSubWallet() {

		}

		nlohmann::json
		IDChainSubWallet::CreateIDTransaction(const std::string &fromAddress, const nlohmann::json &payloadJson,
											  const nlohmann::json &programJson, const std::string &memo) {
			Log::preinfo("{}:{} {} | {} | {} | {} | {}", _parent->GetWalletID(), _info->GetChainID(), GetFun(), fromAddress, payloadJson.dump(), programJson.dump(), memo);

			std::string toAddress;
			Program program;
			PayloadPtr payload = nullptr;
			try {
				toAddress = payloadJson["ID"].get<std::string>();
				program.FromJson(programJson);
				payload = PayloadPtr(new PayloadRegisterIdentification());
				payload->FromJson(payloadJson, 0);
			} catch (const nlohmann::detail::exception &e) {
				ErrorChecker::ThrowParamException(Error::JsonFormatError,
												  "Create id tx param error: " + std::string(e.what()));
			}

			std::vector<TransactionOutput> outputs;
			Address receiveAddr(toAddress);
			outputs.emplace_back(0, receiveAddr, Asset::GetELAAssetID());

			TransactionPtr tx = CreateTx(fromAddress, outputs, memo);

			tx->SetTransactionType(Transaction::RegisterIdentification, payload);

			tx->AddProgram(program);

			nlohmann::json txJson = tx->ToJson();
			Log::retinfo("{}:{} {} | {}", _parent->GetWalletID(), _info->GetChainID(), GetFun(), txJson.dump());
			return txJson;
		}

		void IDChainSubWallet::onTxAdded(const TransactionPtr &transaction) {
			if (transaction != nullptr && transaction->GetTransactionType() == Transaction::RegisterIdentification) {
				std::string txHash = transaction->GetHash().GetHex();

				std::for_each(_callbacks.begin(), _callbacks.end(),
							  [&transaction, &txHash](ISubWalletCallback *callback) {
								  const PayloadRegisterIdentification *payload = static_cast<const PayloadRegisterIdentification *>(
									  transaction->GetPayload());
								  callback->OnTransactionStatusChanged(txHash, "Added", payload->ToJson(0), 0);
							  });
			} else {
				SubWallet::onTxAdded(transaction);
			}
		}

		void IDChainSubWallet::onTxUpdated(const std::vector<uint256> &hashes, uint32_t blockHeight, time_t timeStamp) {
			for (size_t i = 0; i < hashes.size(); ++i) {
				TransactionPtr transaction = _walletManager->getWallet()->TransactionForHash(hashes[i]);
				if (transaction != nullptr &&
					transaction->GetTransactionType() == Transaction::RegisterIdentification) {
					uint32_t confirm = blockHeight >= transaction->GetBlockHeight() ? blockHeight -
																					  transaction->GetBlockHeight() + 1
																					: 0;

					std::for_each(_callbacks.begin(), _callbacks.end(),
								  [&i, &hashes, &confirm, &transaction, this](ISubWalletCallback *callback) {

									  const PayloadRegisterIdentification *payload = static_cast<const PayloadRegisterIdentification *>(
										  transaction->GetPayload());
									  callback->OnTransactionStatusChanged(hashes[i].GetHex(), "Updated", payload->ToJson(0),
																		   confirm);
								  });
				} else {
					SubWallet::onTxUpdated(hashes, blockHeight, timeStamp);
				}
			}
		}

	}
}
