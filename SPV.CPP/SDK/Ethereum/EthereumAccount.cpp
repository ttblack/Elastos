/*
 * EthereumAccount
 *
 * Created by Ed Gamble <ed@breadwallet.com> on 3/20/18.
 * Copyright (c) 2018 Breadwinner AG.  All right reserved.
 * Copyright (c) 2019 Elastos Foundation
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#include "EthereumAccount.h"
#include "EthereumEWM.h"
#include <ethereum/ewm/BREthereumEWM.h>

namespace Elastos {
	namespace ElaWallet {

		EthereumAccount::EthereumAccount(const EthereumEWMPtr &ewm, BREthereumAccount account) :
			Reference(ewm, account) {

		}

		EthereumAccount::~EthereumAccount() {

		}

		std::string EthereumAccount::getPrimaryAddress() const {
			return ewmGetAccountPrimaryAddress(GetEWM()->getRaw());
		}

		bytes_t EthereumAccount::getPrimaryAddressPublicKey() const {
			BRKey key = ewmGetAccountPrimaryAddressPublicKey(GetEWM()->getRaw());
			bytes_t pubkey(key.pubKey, sizeof(key.pubKey));
			return pubkey;
		}

		bytes_t EthereumAccount::getPrimaryAddressPrivateKey(const std::string &paperKey) const {
			BRKey key = ewmGetAccountPrimaryAddressPrivateKey(GetEWM()->getRaw(), paperKey.data());
			bytes_t prvkey(&key, sizeof(key));
			return prvkey;
		}

	}
}