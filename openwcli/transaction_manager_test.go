package openwcli

import (
	"encoding/json"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/go-openw-sdk/openwsdk"
	"github.com/google/uuid"
	"testing"
)

func testFindAccountByID(accountID string, list []*openwsdk.Account) *openwsdk.Account {
	for _, a := range list {
		if a.AccountID == accountID {
			return a
		}
	}
	return nil
}

func TestCLI_Transfer_BTC(t *testing.T) {
	cli := getTestOpenwCLI()
	if cli == nil {
		return
	}

	wallets, err := cli.GetWalletsOnServer()
	if err != nil {
		log.Error("GetWalletsOnServer error:", err)
		return
	}

	if len(wallets) > 0 {
		accounts, err := cli.GetAccountsOnServer(wallets[0].WalletID)
		if err != nil {
			log.Error("GetAccountsOnServer error:", err)
			return
		}

		account := testFindAccountByID("J3wiDj2jMGdp9aqmALhQtEkJQch4YN9e38TEXzRgZyKY", accounts)

		if account != nil {
			sid := uuid.New().String()
			_, _, err = cli.Transfer(wallets[0], accounts[0], "", "mp1JDsi7Dr2PkcWu1j4SUSTXJqXjFMaeVx", "0.023", sid, "", "", "12345678")
			if err != nil {
				log.Error("Transfer error:", err)
				return
			}
		}
	}
}

func TestCLI_Summary_BTC(t *testing.T) {

	cli := getTestOpenwCLI()
	if cli == nil {
		return
	}

	accountID := "J3wiDj2jMGdp9aqmALhQtEkJQch4YN9e38TEXzRgZyKY"

	plain := `

{
	"wallets": [{
		"walletID": "VzRF939isEwpz7wLUwqULpmhct2wsApdm4",
		"password": "12345678",
		"accounts": [{
			"accountID": "J3wiDj2jMGdp9aqmALhQtEkJQch4YN9e38TEXzRgZyKY"
		}]
	}]
}

`
	var summaryTask openwsdk.SummaryTask
	err := json.Unmarshal([]byte(plain), &summaryTask)
	if err != nil {
		log.Error("json.Unmarshal error:", err)
		return
	}

	cli.summaryTask = &summaryTask



	err = cli.SetSummaryInfo(&openwsdk.SummarySetting{
		"VzRF939isEwpz7wLUwqULpmhct2wsApdm4",
		accountID,
		"mxoCkSBmiLQ86N73kXNLHEUgcUBoKdFawH",
		"0.2",
		"0.001",
		"0",
		0,
	})
	if err != nil {
		log.Error("SetSummaryInfo error:", err)
		return
	}

	wallets, err := cli.GetWalletsOnServer()
	if err != nil {
		log.Error("GetWalletsOnServer error:", err)
		return
	}

	if len(wallets) > 0 {
		accounts, err := cli.GetAccountsOnServer(wallets[0].WalletID)
		if err != nil {
			log.Error("GetAccountsOnServer error:", err)
			return
		}

		account := testFindAccountByID(accountID, accounts)

		if account != nil {
			cli.SummaryTask()
		}
	}
}

func TestCLI_GetSummaryTaskLog(t *testing.T) {
	cli := getTestOpenwCLI()
	if cli == nil {
		return
	}
	logs, err := cli.GetSummaryTaskLog(0, 20)
	if err != nil {
		log.Error("GetAccountsOnServer error:", err)
		return
	}
	for i, l := range logs {
		log.Infof("log[%d]: %+v", i, l)
	}
}