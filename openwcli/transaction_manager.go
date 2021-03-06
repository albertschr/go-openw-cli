package openwcli

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/go-openw-sdk/openwsdk"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/owtp"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

//GetTokenBalance 获取代币余额
func (cli *CLI) GetTokenBalance(account *openwsdk.Account, contractID string) string {
	getBalance := "0"
	cli.api.GetTokenBalanceByAccount(account.AccountID, contractID, true,
		func(status uint64, msg string, balance *openwsdk.TokenBalance) {
			if status == owtp.StatusSuccess {
				getBalance = balance.Balance.Balance
			}
		})
	return getBalance
}

//Transfer 转账交易
func (cli *CLI) Transfer(wallet *openwsdk.Wallet, account *openwsdk.Account, contractAddress, to, amount, sid, feeRate, memo, password string) ([]*openwsdk.Transaction, []*openwsdk.FailedRawTransaction, error) {

	var (
		isContract  bool
		retTx       []*openwsdk.Transaction
		retFailed   []*openwsdk.FailedRawTransaction
		retRawTx    *openwsdk.RawTransaction
		err         error
		createErr   error
		contractID  string
		tokenSymbol string
	)

	//获取种子文件
	key, err := cli.getLocalKeyByWallet(wallet, password)
	if err != nil {
		return nil, nil, err
	}

	if len(contractAddress) > 0 {
		isContract = true
		token, findErr := cli.GetTokenContractList("Symbol", account.Symbol, "Address", contractAddress)
		if findErr != nil {
			return nil, nil, findErr
		}
		contractID = token[0].ContractID
		tokenSymbol = token[0].Token
	}
	coin := openwsdk.Coin{
		Symbol:     account.Symbol,
		IsContract: isContract,
		ContractID: contractID,
	}

	api := cli.api
	err = api.CreateTrade(account.AccountID, sid, coin, amount, to, feeRate, memo, true,
		func(status uint64, msg string, rawTx *openwsdk.RawTransaction) {
			if status != owtp.StatusSuccess {
				createErr = fmt.Errorf(msg)
				return
			}
			retRawTx = rawTx
		})
	if err != nil {
		return nil, nil, err
	}
	if createErr != nil {
		return nil, nil, createErr
	}

	//:打印交易单明细
	log.Infof("-----------------------------------------------")
	log.Infof("[%s %s Transfer]", account.Symbol, tokenSymbol)
	log.Infof("From Account: %s", account.AccountID)
	log.Infof("To Address: %s", to)
	log.Infof("Send Amount: %s", amount)
	log.Infof("Fees: %v", retRawTx.Fees)
	log.Infof("FeeRate: %v", retRawTx.FeeRate)
	log.Infof("-----------------------------------------------")

	//签名交易单
	err = openwsdk.SignRawTransaction(retRawTx, key)
	if err != nil {
		return nil, nil, err
	}

	//广播交易单
	err = api.SubmitTrade([]*openwsdk.RawTransaction{retRawTx}, true,
		func(status uint64, msg string, successTx []*openwsdk.Transaction, failedRawTxs []*openwsdk.FailedRawTransaction) {
			if status != owtp.StatusSuccess {
				createErr = fmt.Errorf(msg)
				return
			}

			retTx = successTx
			retFailed = failedRawTxs
		})
	if err != nil {
		return nil, nil, err
	}
	if createErr != nil {
		return nil, nil, createErr
	}

	if len(retTx) > 0 {
		//打印交易单
		log.Info("send transaction successfully.")
		log.Info("transaction id:", retTx[0].Txid)
	} else if len(retFailed) > 0 {
		//打印交易单
		log.Error("send transaction failed. unexpected error:", retFailed[0].Reason)
	}

	return retTx, retFailed, nil
}

//SummaryWallets 执行汇总流程
func (cli *CLI) SummaryTask() {

	log.Infof("[Summary Task Start]------%s", common.TimeFormat("2006-01-02 15:04:05"))

	cli.mu.RLock()
	defer cli.mu.RUnlock()
	//读取参与汇总的钱包
	for _, task := range cli.summaryTask.Wallets {

		if task.Wallet == nil {
			w, err := cli.GetWalletByWalletIDOnLocal(task.WalletID)
			if err != nil {
				log.Errorf("Summary wallet[%s] unexpected error: %v", task.WalletID, err)
				continue
			}
			task.Wallet = w
		}

		key, err := cli.getLocalKeyByWallet(task.Wallet, task.Password)
		if err != nil {
			log.Errorf("Summary wallet[%s] unexpected error: %v", task.WalletID, err)
			continue
		}

		for _, accountTask := range task.Accounts {

			account, err := cli.GetAccountByAccountID(accountTask.AccountID)
			if err != nil {
				continue
			}

			//汇总账户主币
			err = cli.SummaryAccountMainCoin(accountTask, account, key)
			if err != nil {
				log.Errorf("Summary wallet[%s] account[%s] main coin unexpected error: %v", task.WalletID, account.AccountID, err)
			}

			//汇总账户主币
			err = cli.SummaryAccountTokenContracts(accountTask, account, key)
			if err != nil {
				log.Errorf("Summary wallet[%s] account[%s] token contracts unexpected error: %v", task.WalletID, account.AccountID, err)
			}

		}

	}

	log.Infof("[Summary Task End]------%s", common.TimeFormat("2006-01-02 15:04:05"))
}

//SummaryAccountMainCoin 汇总账户主币
func (cli *CLI) SummaryAccountMainCoin(accountTask *openwsdk.SummaryAccountTask, account *openwsdk.Account, key *hdkeystore.HDKey) error {

	var (
		err     error
		sumSets openwsdk.SummarySetting
	)

	//读取汇总信息
	err = cli.db.One("AccountID", account.AccountID, &sumSets)
	if err != nil {
		return fmt.Errorf("Summary account[%s] can not find account summary setting ", account.AccountID)
	}

	if sumSets.SumAddress == "" {
		log.Errorf("Summary account[%s] summary address is empty!", account.AccountID)
		return err
	}

	//balance, _ := decimal.NewFromString(account.Balance)
	//threshold, _ := decimal.NewFromString(sumSets.Threshold)

	coin := openwsdk.Coin{
		Symbol:     account.Symbol,
		IsContract: false,
	}

	log.Infof("Summary account[%s] Symbol: %s start", account.AccountID, account.Symbol)

	err = cli.summaryAccountProcess(account, accountTask, key, account.Balance, sumSets, coin)

	log.Infof("Summary account[%s] Symbol: %s end", account.AccountID, account.Symbol)

	if err != nil {
		return err
	}

	return nil
}

//SummaryAccountTokenContracts 汇总账户代币合约
func (cli *CLI) SummaryAccountTokenContracts(accountTask *openwsdk.SummaryAccountTask, account *openwsdk.Account, key *hdkeystore.HDKey) error {

	var (
		err     error
		sumSets openwsdk.SummarySetting
	)

	if len(accountTask.Contracts) == 0 {
		return nil
	}

	//tokens, err := cli.GetTokenContractList("Symbol", account.Symbol)
	//if err != nil {
	//	return err
	//}

	tokenBalances, err := cli.GetAllTokenContractBalance(account.AccountID)
	if err != nil {
		return err
	}

	//查询已选token过程
	findSelectedTokensFunc := func(t string) bool {

		if accountTask.Contracts[0] == "all" {
			return true
		}

		for _, c := range accountTask.Contracts {
			if c == t {
				return true
			}
		}
		return false
	}

	//读取汇总信息
	err = cli.db.One("AccountID", account.AccountID, &sumSets)
	if err != nil {
		return err
	}

	for _, token := range tokenBalances {

		//找不到已选合约跳到下一个
		if !findSelectedTokensFunc(token.Address) {
			continue
		}

		if sumSets.SumAddress == "" {
			log.Errorf("Summary account[%s] summary address is empty!")
			return err
		}

		//查询合约余额
		tokenBalance := cli.GetTokenBalance(account, token.ContractID)

		coin := openwsdk.Coin{
			Symbol:     account.Symbol,
			IsContract: true,
			ContractID: token.ContractID,
		}

		log.Infof("Summary account[%s] Symbol: %s, token: %s start", account.AccountID, account.Symbol, token.Token)

		err = cli.summaryAccountProcess(account, accountTask, key, tokenBalance, sumSets, coin)

		log.Infof("Summary account[%s] Symbol: %s, token: %s end", account.AccountID, account.Symbol, token.Token)

		if err != nil {
			continue
		}

	}
	return nil
}

//summaryAccountProcess 汇总账户过程
func (cli *CLI) summaryAccountProcess(account *openwsdk.Account, task *openwsdk.SummaryAccountTask, key *hdkeystore.HDKey, balance string, sumSets openwsdk.SummarySetting, coin openwsdk.Coin) error {

	const (
		limit = 200
	)

	var (
		err       error
		createErr error
		retTx     []*openwsdk.Transaction
		retFailed []*openwsdk.FailedRawTransaction
		retRawTxs []*openwsdk.RawTransaction
	)

	balanceDec, _ := decimal.NewFromString(balance)
	threshold, _ := decimal.NewFromString(sumSets.Threshold)
	//如果余额大于阀值，汇总的地址
	if balanceDec.GreaterThan(threshold) {

		log.Infof("Summary account[%s] Current Balance = %v ", account.AccountID, balance)
		log.Infof("Summary account[%s] Summary Address = %v ", account.AccountID, sumSets.SumAddress)
		log.Infof("Summary account[%s] Start Create Summary Transaction", account.AccountID)

		//分页汇总交易
		for i := 0; i < int(account.AddressIndex)+1; i = i + limit {
			err = nil
			retRawTxs = nil
			retTx = nil
			retFailed = nil

			log.Infof("Create Summary Transaction in address range [%d...%d]", i, i+limit)

			//:记录汇总批次号
			sid := uuid.New().String()

			err = cli.api.CreateSummaryTx(account.AccountID, sumSets.SumAddress, coin,
				task.FeeRate, sumSets.MinTransfer, sumSets.RetainedBalance,
				i, limit, sumSets.Confirms, sid, true,
				func(status uint64, msg string, rawTxs []*openwsdk.RawTransaction) {
					retRawTxs = rawTxs
					if status != owtp.StatusSuccess {
						createErr = fmt.Errorf(msg)
					}
				})

			if err != nil {
				log.Warn("CreateSummaryTransaction unexpected error: %v", err)
				continue
			}

			if createErr != nil {
				log.Warn("CreateSummaryTransaction unexpected error:", createErr)
				continue
			}

			signedRawTxs := make([]*openwsdk.RawTransaction, 0)
			txIDs := make([]string, 0)
			sids := make([]string, 0)
			for _, rawTx := range retRawTxs {
				//签名交易
				err = openwsdk.SignRawTransaction(rawTx, key)
				if err != nil {
					log.Warn("SignRawTransaction unexpected error: %v", err)
					continue
				}

				signedRawTxs = append(signedRawTxs, rawTx)
			}

			if len(signedRawTxs) == 0 {
				continue
			}

			//	广播交易单
			err = cli.api.SubmitTrade(signedRawTxs, true,
				func(status uint64, msg string, successTx []*openwsdk.Transaction, failedRawTxs []*openwsdk.FailedRawTransaction) {
					if status != owtp.StatusSuccess {
						createErr = fmt.Errorf(msg)
						return
					}

					retTx = successTx
					retFailed = failedRawTxs
				})
			if err != nil {
				log.Warningf("SubmitRawTransaction unexpected error: %v", err)
				continue
			}
			if createErr != nil {
				log.Warningf("SubmitRawTransaction unexpected error: %v", createErr)
				continue
			}

			//打印汇总交易结果
			totalSumAmount := decimal.Zero
			totalCostFees := decimal.Zero

			for _, tx := range retTx {
				log.Infof("[Success] txid: %s", tx.Txid)
				//:计算总的汇总数量，手续费

				fees, _ := decimal.NewFromString(tx.Fees)

				totalCostFees = totalCostFees.Add(fees)
				txIDs = append(txIDs, tx.Txid)
				sids = append(sids, tx.Sid)
				//统计汇总总数
				for i, a := range tx.ToAddress {
					if a == sumSets.SumAddress {
						amount, _ := decimal.NewFromString(tx.ToAddressV[i])
						totalSumAmount = totalSumAmount.Add(amount)
					}
				}
			}

			for _, tx := range retFailed {
				log.Warn("[Failed] reason:", tx.Reason)
			}

			//:记录汇总情况
			totalSumAmount = totalSumAmount.Sub(totalCostFees)
			summaryTaskLog := openwsdk.SummaryTaskLog{
				Sid:            sid,
				WalletID:       account.WalletID,
				AccountID:      account.AccountID,
				StartAddrIndex: i,
				EndAddrIndex:   i + limit,
				Coin:           coin,
				SuccessCount:   len(retTx),
				FailCount:      len(retFailed),
				TxIDs:          txIDs,
				Sids:           sids,
				TotalSumAmount: totalSumAmount.String(),
				TotalCostFees:  totalCostFees.String(),
				CreateTime:     time.Now().Unix(),
			}
			err = cli.db.Save(&summaryTaskLog)
			if err != nil {
				log.Infof("Save summary task log failed: %s", err.Error())
			} else {
				log.Infof("Save summary task log successfully")
			}

		}
	} else {
		log.Infof("Summary account[%s] Current Balance: %v, below threshold: %v", account.AccountID, balance, threshold)
	}

	return nil
}

func FindExistedSummaryWalletTask(walletID string, tasks []*openwsdk.SummaryWalletTask) (int, *openwsdk.SummaryWalletTask) {
	for i, w := range tasks {
		if w.WalletID == walletID {
			return i, w
		}
	}
	return -1, nil
}

func FindExistedSummaryAccountTask(accountID string, tasks []*openwsdk.SummaryAccountTask) (int, *openwsdk.SummaryAccountTask) {
	for i, w := range tasks {
		if w.AccountID == accountID {
			return i, w
		}
	}
	return -1, nil
}

//checkSummaryTaskIsHaveSettings 检查汇总任务中的账户是否已配置
func (cli *CLI) checkSummaryTaskIsHaveSettings(task *openwsdk.SummaryTask) error {

	for _, w := range task.Wallets {
		for _, account := range w.Accounts {
			var sumSets openwsdk.SummarySetting
			//读取汇总信息
			err := cli.db.One("AccountID", account.AccountID, &sumSets)
			if err != nil {
				return fmt.Errorf("Summary account[%s] can not find account summary setting ", account.AccountID)
			}

			if sumSets.SumAddress == "" {
				log.Errorf("Summary account[%s] summary address is empty!", account.AccountID)
				return err
			}

			account.SummarySetting = &sumSets
		}
	}
	return nil
}

func (cli *CLI) appendSummaryWalletTasks(sums []*openwsdk.SummaryWalletTask) {
	cli.mu.Lock()
	defer cli.mu.Unlock()

	for _, newWalletTask := range sums {

		//查找钱包是否汇总中
		_, executingWallet := FindExistedSummaryWalletTask(newWalletTask.WalletID, cli.summaryTask.Wallets)
		if executingWallet != nil {
			//钱包汇总中...
			for _, newAccountTask := range newWalletTask.Accounts {
				//查找账户是否汇总中
				_, executingAccount := FindExistedSummaryAccountTask(newAccountTask.AccountID, executingWallet.Accounts)
				if executingAccount != nil {
					//账户汇总中...
					//重置汇总的合约
					executingAccount.Contracts = newAccountTask.Contracts
				} else {
					executingWallet.Accounts = append(executingWallet.Accounts, newAccountTask)
					log.Infof("Summary account[%s] task has been appended ", newAccountTask.AccountID)
				}
			}

		} else {
			cli.summaryTask.Wallets = append(cli.summaryTask.Wallets, newWalletTask)
			log.Infof("Summary wallet[%s] task has been appended ", newWalletTask.WalletID)
		}
	}

}

func (cli *CLI) removeSummaryWalletTasks(walletID string, accountID string) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	indexWallet, executingWallet := FindExistedSummaryWalletTask(walletID, cli.summaryTask.Wallets)
	if executingWallet != nil {
		if len(accountID) > 0 {
			//查找账户是否汇总中
			indexAccount, executingAccount := FindExistedSummaryAccountTask(accountID, executingWallet.Accounts)
			if executingAccount != nil {
				//移除汇总账户任务
				executingWallet.Accounts = append(executingWallet.Accounts[:indexAccount], executingWallet.Accounts[indexAccount+1:]...)
				log.Infof("Summary account[%s] task has been removed ", accountID)
			}
		} else {
			//移除汇总钱包任务
			cli.summaryTask.Wallets = append(cli.summaryTask.Wallets[:indexWallet], cli.summaryTask.Wallets[indexWallet+1:]...)
			log.Infof("Summary wallet[%s] task has been removed ", walletID)
		}
	}
}

func (cli *CLI) GetSummaryTaskLog(offset, limit int64) ([]*openwsdk.SummaryTaskLog, error) {
	var summaryTaskLog []*openwsdk.SummaryTaskLog
	//err := cli.db.All(&summaryTaskLog)
	err := cli.db.AllByIndex("CreateTime", &summaryTaskLog,
		storm.Limit(int(limit)), storm.Skip(int(offset)), storm.Reverse())
	if err != nil {
		return nil, err
	}
	return summaryTaskLog, nil
}
