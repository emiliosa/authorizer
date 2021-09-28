package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Account struct {
	ActiveCard     bool `json:"active-card"`
	AvailableLimit int  `json:"available-limit"`
}

type Transaction struct {
	Merchant string      `json:"merchant"`
	Amount   int         `json:"amount"`
	Time     interface{} `json:"time"`
}

type AccountOperation struct {
	Account Account
}

type AccountOperationOutput struct {
	Account    Account  `json:"account"`
	Violations []string `json:"violations"`
}

type TransactionOperation struct {
	Transaction Transaction
}

const (
	ACCOUNT_NOT_INITIALIZED       = "account-not-initialized"
	ACCOUNT_ALREADY_INITIALIZED   = "account-already-initialized"
	CARD_NOT_ACTIVE               = "card-not-active"
	DOUBLED_TRANSACTION           = "doubled-transaction"
	INSUFFICIENT_LIMIT            = "insufficient-limit"
	HIGH_FRECUENCY_SMALL_INTERVAL = "high-frequency-small-interval"
)

func main() {
	// Open our jsonFile
	file, err := os.Open("operations.txt")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	var accountOperationsOutput []AccountOperationOutput
	var transactionOperations []TransactionOperation
	var account Account
	var initialAccount Account
	var output AccountOperationOutput
	var hasAccount = false

	scanner := bufio.NewScanner(file)

	// defer the closing of our file so that we can parse it later on
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}(file)

	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()

		var result map[string]interface{}
		json.Unmarshal([]byte(line), &result)

		//fmt.Println("Text line:", line, "unmarshaled:", result)

		switch true {
		case result["account"] != nil:
			var accountOperation AccountOperation
			var violations []string

			json.Unmarshal([]byte(line), &accountOperation)
			//fmt.Println("JSON parse[account]: ", accountOperation.Account.ActiveCard, accountOperation.Account.AvailableLimit)

			if !hasAccount {
				account, initialAccount = accountOperation.Account, accountOperation.Account
				hasAccount = true
				output = AccountOperationOutput{Account{
					ActiveCard:     accountOperation.Account.ActiveCard,
					AvailableLimit: accountOperation.Account.AvailableLimit,
				}, violations}
			} else {
				violations = []string{ACCOUNT_ALREADY_INITIALIZED}
				output = AccountOperationOutput{Account{
					ActiveCard:     account.ActiveCard,
					AvailableLimit: account.AvailableLimit,
				}, violations}
			}

			accountOperationsOutput = append(accountOperationsOutput, output)
		case result["transaction"] != nil:
			var transactionOperation TransactionOperation
			var violations []string

			if !hasAccount {
				violations = []string{ACCOUNT_NOT_INITIALIZED}
			} else {
				json.Unmarshal([]byte(line), &transactionOperation)
				//fmt.Println("JSON parse[transaction]: ", transactionOperation)

				transactionOperation.Transaction.Time, _ = parseTime(transactionOperation.Transaction.Time.(string))

				var availableLimit = account.AvailableLimit - transactionOperation.Transaction.Amount

				if availableLimit < 0 {
					violations = []string{INSUFFICIENT_LIMIT}
				}

				if !account.ActiveCard {
					violations = append(violations, CARD_NOT_ACTIVE)
				}

				if hasDoubledTransaction(transactionOperations, transactionOperation.Transaction) {
					violations = append(violations, DOUBLED_TRANSACTION)
				}

				if hasHighFrecuencySmallInterval(transactionOperations, transactionOperation.Transaction, initialAccount) {
					violations = append(violations, HIGH_FRECUENCY_SMALL_INTERVAL)
				}

				if violations == nil {
					account.AvailableLimit = availableLimit
				}
			}

			output = AccountOperationOutput{Account{
				ActiveCard:     account.ActiveCard,
				AvailableLimit: account.AvailableLimit,
			}, violations}

			accountOperationsOutput = append(accountOperationsOutput, output)
			transactionOperations = append(transactionOperations, transactionOperation)
		default:
			fmt.Println("operation not valid")
		}
	}

	for _, item := range accountOperationsOutput {
		data, _ := json.Marshal(item)
		fmt.Println("accountOperationsOutput:", string(data))
	}
}

func parseTime(data string) (time.Time, error) {
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	return time.ParseInLocation(time.RFC3339, data, loc)
}

func hasDoubledTransaction(slice []TransactionOperation, transaction Transaction) bool {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i].Transaction.Merchant == transaction.Merchant && slice[i].Transaction.Amount == transaction.Amount {
			if transaction.Time.(time.Time).Sub(slice[i].Transaction.Time.(time.Time)).Minutes() < 2 {
				return true
			}
		}
	}
	return false
}

func hasHighFrecuencySmallInterval(slice []TransactionOperation, transaction Transaction, account Account) bool {
	if account.ActiveCard && account.AvailableLimit == 100 {
		for count, i := 0, len(slice)-1; i >= 0; i-- {
			if transaction.Time.(time.Time).Sub(slice[i].Transaction.Time.(time.Time)).Minutes() < 2 {
				count++
				if count == 3 {
					return true
				}
			}
		}
	}
	return false
}