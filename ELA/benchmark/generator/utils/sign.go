package utils

import (
	"github.com/elastos/Elastos.ELA/account"
	"github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/contract/program"
	"github.com/elastos/Elastos.ELA/core/types"
)

func SignStandardTx(tx *types.Transaction, ac *account.Account) (err error) {
	accounts := map[common.Uint160]*account.Account{}
	accounts[ac.ProgramHash.ToCodeHash()] = ac

	pg := &program.Program{
		Code: ac.RedeemScript,
	}
	pg, err = account.SignStandardTransaction(tx, pg, accounts)
	if err != nil {
		return
	}

	tx.Programs = []*program.Program{pg}
	return
}
