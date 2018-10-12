// Copyright (c) 2012-2018 The Elastos Open Source Project
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

#include <sstream>

#include "Common/Log.h"
#include "CoreWalletManager.h"
#include "Account/SingleSubAccount.h"

namespace Elastos {
	namespace ElaWallet {

		CoreWalletManager::CoreWalletManager(const PluginTypes &pluginTypes, const ChainParams &chainParams) :
				PeerManager::Listener(pluginTypes),
				_wallet(nullptr),
				_walletListener(nullptr),
				_peerManager(nullptr),
				_peerManagerListener(nullptr),
				_subAccount(nullptr),
				_pluginTypes(pluginTypes),
				_chainParams(chainParams) {
		}

		CoreWalletManager::~CoreWalletManager() {

		}

		void CoreWalletManager::init(const SubAccountPtr &subAccount, uint32_t earliestPeerTime, uint32_t reconnectSeconds) {
			_subAccount = subAccount;
			_earliestPeerTime = earliestPeerTime;
			_reconnectSeconds = reconnectSeconds;
		}

		const WalletPtr &CoreWalletManager::getWallet() {
			if (_wallet == nullptr) {
				_wallet = WalletPtr(new Wallet(loadTransactions(), _subAccount, createWalletListener()));
			}
			return _wallet;
		}

		const PeerManagerPtr &CoreWalletManager::getPeerManager() {
			if (_peerManager == nullptr) {
				_peerManager = PeerManagerPtr(new PeerManager(
						_chainParams,
						getWallet(),
						_earliestPeerTime,
						_reconnectSeconds,
						loadBlocks(),
						loadPeers(),
						createPeerManagerListener(),
						_pluginTypes));
			}

			return _peerManager;
		}

		void CoreWalletManager::balanceChanged(uint64_t balance) {

		}

		void CoreWalletManager::onTxAdded(const TransactionPtr &transaction) {

		}

		void CoreWalletManager::onTxUpdated(const std::string &hash, uint32_t blockHeight, uint32_t timeStamp) {

		}

		void CoreWalletManager::onTxDeleted(const std::string &hash, bool notifyUser, bool recommendRescan) {

		}

		void CoreWalletManager::syncStarted() {

		}

		void CoreWalletManager::syncStopped(const std::string &error) {

		}

		void CoreWalletManager::txStatusUpdate() {

		}

		void
		CoreWalletManager::saveBlocks(bool replace, const std::vector<MerkleBlockPtr> &blocks) {

		}

		void CoreWalletManager::savePeers(bool replace, const std::vector<PeerPtr> &peers) {

		}

		bool CoreWalletManager::networkIsReachable() {
			return true;
		}

		void CoreWalletManager::txPublished(const std::string &error) {

		}

		void CoreWalletManager::blockHeightIncreased(uint32_t blockHeight) {

		}

		std::vector<TransactionPtr> CoreWalletManager::loadTransactions() {
			//todo complete me
			return std::vector<TransactionPtr>();
		}

		std::vector<MerkleBlockPtr> CoreWalletManager::loadBlocks() {
			//todo complete me
			return std::vector<MerkleBlockPtr>();
		}

		std::vector<PeerInfo> CoreWalletManager::loadPeers() {
			//todo complete me
			return std::vector<PeerInfo>();
		}

		int CoreWalletManager::getForkId() const {
			//todo complete me
			return -1;
		}

		const CoreWalletManager::PeerManagerListenerPtr &CoreWalletManager::createPeerManagerListener() {
			if (_peerManagerListener == nullptr) {
				_peerManagerListener = PeerManagerListenerPtr(
						new WrappedExceptionPeerManagerListener(this, _pluginTypes));
			}
			return _peerManagerListener;
		}

		const CoreWalletManager::WalletListenerPtr &CoreWalletManager::createWalletListener() {
			if (_walletListener == nullptr) {
				_walletListener = WalletListenerPtr(new WrappedExceptionWalletListener(this));
			}
			return _walletListener;
		}

		WrappedExceptionPeerManagerListener::WrappedExceptionPeerManagerListener(PeerManager::Listener *listener,
																				 const PluginTypes &pluginTypes) :
				PeerManager::Listener(pluginTypes),
				_listener(listener) {
		}

		void WrappedExceptionPeerManagerListener::syncStarted() {
			try {
				_listener->syncStarted();
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (syncStarted) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (syncStarted) error.");
			}
		}

		void WrappedExceptionPeerManagerListener::syncStopped(const std::string &error) {
			try {
				_listener->syncStopped(error);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (syncStopped) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (syncStopped) error.");
			}
		}

		void WrappedExceptionPeerManagerListener::txStatusUpdate() {
			try {
				_listener->txStatusUpdate();
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (txStatusUpdate) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (txStatusUpdate) error.");
			}
		}

		void WrappedExceptionPeerManagerListener::saveBlocks(bool replace, const std::vector<MerkleBlockPtr> &blocks) {

			try {
				_listener->saveBlocks(replace, blocks);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (saveBlocks) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (saveBlocks) error.");
			}
		}

		void
		WrappedExceptionPeerManagerListener::savePeers(bool replace, const std::vector<PeerPtr> &peers) {

			try {
				_listener->savePeers(replace, peers);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (savePeers) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (savePeers) error.");
			}
		}

		bool WrappedExceptionPeerManagerListener::networkIsReachable() {
			try {
				return _listener->networkIsReachable();
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (networkIsReachable) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (networkIsReachable) error.");
			}

			return true;
		}

		void WrappedExceptionPeerManagerListener::txPublished(const std::string &error) {
			try {
				_listener->txPublished(error);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (txPublished) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (txPublished) error.");
			}
		}

		void WrappedExceptionPeerManagerListener::blockHeightIncreased(uint32_t blockHeight) {
			try {
				_listener->blockHeightIncreased(blockHeight);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (blockHeightIncreased) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (blockHeightIncreased) error.");
			}
		}

		void WrappedExceptionPeerManagerListener::syncIsInactive(uint32_t time) {
			try {
				_listener->syncIsInactive(time);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Peer manager callback (blockHeightIncreased) error: {}", ex.what());
			}
			catch (...) {
				Log::getLogger()->error("Peer manager callback (blockHeightIncreased) error.");
			}
		}

		WrappedExecutorPeerManagerListener::WrappedExecutorPeerManagerListener(
				PeerManager::Listener *listener,
				Executor *executor,
				Executor *reconnectExecutor,
				const PluginTypes &pluginTypes) :
				PeerManager::Listener(pluginTypes),
				_listener(listener),
				_executor(executor),
				_reconnectExecutor(reconnectExecutor) {
		}

		void WrappedExecutorPeerManagerListener::syncStarted() {
			_executor->execute(Runnable([this]() -> void {
				try {
					_listener->syncStarted();
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (syncStarted) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (syncStarted) error.");
				}
			}));
		}

		void WrappedExecutorPeerManagerListener::syncStopped(const std::string &error) {
			_executor->execute(Runnable([this, error]() -> void {
				try {
					_listener->syncStopped(error);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (syncStopped) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (syncStopped) error.");
				}
			}));
		}

		void WrappedExecutorPeerManagerListener::txStatusUpdate() {
			_executor->execute(Runnable([this]() -> void {
				try {
					_listener->txStatusUpdate();
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (txStatusUpdate) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (txStatusUpdate) error.");
				}
			}));
		}

		void WrappedExecutorPeerManagerListener::saveBlocks(bool replace, const std::vector<MerkleBlockPtr> &blocks) {
			_executor->execute(Runnable([this, replace, &blocks]() -> void {
				try {
					_listener->saveBlocks(replace, blocks);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (saveBlocks) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (saveBlocks) error.");
				}
			}));
		}

		void
		WrappedExecutorPeerManagerListener::savePeers(bool replace, const std::vector<PeerPtr> &peers) {
			_executor->execute(Runnable([this, replace, &peers]() -> void {
				try {
					_listener->savePeers(replace, peers);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (savePeers) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (savePeers) error.");
				}
			}));
		}

		bool WrappedExecutorPeerManagerListener::networkIsReachable() {

			bool result = true;
			_executor->execute(Runnable([this, result]() -> void {
				try {
					_listener->networkIsReachable();
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (networkIsReachable) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (networkIsReachable) error.");
				}
			}));
			return result;
		}

		void WrappedExecutorPeerManagerListener::txPublished(const std::string &error) {

			_executor->execute(Runnable([this, error]() -> void {
				try {
					_listener->txPublished(error);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (txPublished) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (txPublished) error.");
				}
			}));
		}

		void WrappedExecutorPeerManagerListener::blockHeightIncreased(uint32_t blockHeight) {
			_executor->execute(Runnable([this, blockHeight]() -> void {
				try {
					_listener->blockHeightIncreased(blockHeight);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (blockHeightIncreased) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (blockHeightIncreased) error.");
				}
			}));
		}

		void WrappedExecutorPeerManagerListener::syncIsInactive(uint32_t time) {
			_reconnectExecutor->execute(Runnable([this, time]() -> void {
				try {
					_listener->syncIsInactive(time);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Peer manager callback (blockHeightIncreased) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Peer manager callback (blockHeightIncreased) error.");
				}
			}));
		}

		WrappedExceptionWalletListener::WrappedExceptionWalletListener(Wallet::Listener *listener) :
				_listener(listener) {
		}

		void WrappedExceptionWalletListener::balanceChanged(uint64_t balance) {
			try {
				_listener->balanceChanged(balance);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Wallet callback (balanceChanged) error: {}", ex.what());
			}
			catch (...) {
				Log::error("Wallet callback (balanceChanged) error.");
			}
		}

		void WrappedExceptionWalletListener::onTxAdded(const TransactionPtr &transaction) {
			try {
				return _listener->onTxAdded(transaction);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Wallet callback (onTxAdded) error: {}", ex.what());
			}
			catch (...) {
				Log::error("Wallet callback (onTxAdded) error.");
			}
		}

		void WrappedExceptionWalletListener::onTxUpdated(
				const std::string &hash, uint32_t blockHeight, uint32_t timeStamp) {

			try {
				_listener->onTxUpdated(hash, blockHeight, timeStamp);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Wallet callback (onTxUpdated) error: {}", ex.what());
			}
			catch (...) {
				Log::error("Wallet callback (onTxUpdated) error.");
			}
		}

		void WrappedExceptionWalletListener::onTxDeleted(
				const std::string &hash, bool notifyUser, bool recommendRescan) {

			try {
				_listener->onTxDeleted(hash, notifyUser, recommendRescan);
			}
			catch (std::exception ex) {
				Log::getLogger()->error("Wallet callback (onTxDeleted) error: {}", ex.what());
			}
			catch (...) {
				Log::error("Wallet callback (onTxDeleted) error.");
			}
		}

		WrappedExecutorWalletListener::WrappedExecutorWalletListener(
				Wallet::Listener *listener,
				Executor *executor) :
				_listener(listener),
				_executor(executor) {
		}

		void WrappedExecutorWalletListener::balanceChanged(uint64_t balance) {
			_executor->execute(Runnable([this, balance]() -> void {
				try {
					_listener->balanceChanged(balance);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Wallet callback (balanceChanged) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Wallet callback (balanceChanged) error.");
				}
			}));
		}

		void WrappedExecutorWalletListener::onTxAdded(const TransactionPtr &transaction) {
			_executor->execute(Runnable([this, transaction]() -> void {
				try {
					_listener->onTxAdded(transaction);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Wallet callback (onTxAdded) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Wallet callback (onTxAdded) error.");
				}
			}));
		}

		void WrappedExecutorWalletListener::onTxUpdated(
				const std::string &hash, uint32_t blockHeight, uint32_t timeStamp) {
			_executor->execute(Runnable([this, hash, blockHeight, timeStamp]() -> void {
				try {
					_listener->onTxUpdated(hash, blockHeight, timeStamp);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Wallet callback (onTxUpdated) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Wallet callback (onTxUpdated) error.");
				}
			}));
		}

		void WrappedExecutorWalletListener::onTxDeleted(
				const std::string &hash, bool notifyUser, bool recommendRescan) {
			_executor->execute(Runnable([this, hash, notifyUser, recommendRescan]() -> void {
				try {
					_listener->onTxDeleted(hash, notifyUser, recommendRescan);
				}
				catch (std::exception ex) {
					Log::getLogger()->error("Wallet callback (onTxDeleted) error: {}", ex.what());
				}
				catch (...) {
					Log::error("Wallet callback (onTxDeleted) error.");
				}
			}));
		}
	}
}