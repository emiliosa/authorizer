package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
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

type TransactionOperation struct {
	Transaction Transaction
}

type Operations struct {
	input  []interface{}
	output []AccountOperationOutput
}

type AccountOperationOutput struct {
	Account    Account  `json:"account"`
	Violations []string `json:"violations"`
}

type AccountStatus struct {
	account    Account
	hasAccount bool
}

const (
	AccountNotInitialized      = "account-not-initialized"
	AccountAlreadyInitialized  = "account-already-initialized"
	CardNotActive              = "card-not-active"
	DoubledTransaction         = "doubled-transaction"
	InsufficientLimit          = "insufficient-limit"
	HighFrequencySmallInterval = "high-frequency-small-interval"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	operations := process(scanner)
	out := output(operations)
	fmt.Println(out)
}

func output(slice []AccountOperationOutput) string {
	var out []string
	for i := range slice {
		item := slice[i]
		if item.Violations == nil {
			item.Violations = make([]string, 0)
		}
		jsonData, _ := json.Marshal(item)
		out = append(out, string(jsonData))
	}
	return strings.Join(out[:], "\n")
}

func processAccount(operation AccountOperation, accountStatus AccountStatus) AccountOperationOutput {
	var violations []string
	var activeCard bool
	var availableLimit int

	if accountStatus.hasAccount && accountStatus.account.ActiveCard == operation.Account.ActiveCard {
		violations = []string{AccountAlreadyInitialized}
		activeCard = accountStatus.account.ActiveCard
		availableLimit = accountStatus.account.AvailableLimit
	} else {
		activeCard = operation.Account.ActiveCard
		availableLimit = operation.Account.AvailableLimit
	}

	return AccountOperationOutput{Account{
		ActiveCard:     activeCard,
		AvailableLimit: availableLimit,
	}, violations}
}

func processTransaction(new TransactionOperation, status AccountStatus, operations []interface{}) AccountOperationOutput {
	var violations []string
	var account Account

	if !status.hasAccount {
		account.ActiveCard = false
		account.AvailableLimit = 0
		violations = []string{AccountNotInitialized}
	} else {
		account.ActiveCard = status.account.ActiveCard
		account.AvailableLimit = status.account.AvailableLimit - new.Transaction.Amount

		if account.AvailableLimit < 0 {
			violations = []string{InsufficientLimit}
		}

		if !status.account.ActiveCard {
			violations = append(violations, CardNotActive)
		}

		if hasDoubledTransaction(operations, new.Transaction) {
			violations = append(violations, DoubledTransaction)
		}

		if hasHighFrequencySmallInterval(operations, new.Transaction) {
			violations = append(violations, HighFrequencySmallInterval)
		}

		if violations != nil {
			account.AvailableLimit = status.account.AvailableLimit
		}
	}

	return AccountOperationOutput{account, violations}
}

func process(scanner *bufio.Scanner) []AccountOperationOutput {
	var operations = Operations{}
	var accountStatus = AccountStatus{
		account:    Account{},
		hasAccount: false,
	}

	for scanner.Scan() {
		line := scanner.Text()

		// set interface type for unstructured json
		var result map[string]interface{}
		err := json.Unmarshal([]byte(line), &result)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch true {
		case result["account"] != nil: // check json structure match account structure
			var accountOperation AccountOperation
			err := json.Unmarshal([]byte(line), &accountOperation)
			if err != nil {
				fmt.Println(err)
				continue
			}

			output := processAccount(accountOperation, accountStatus)

			if output.Violations == nil {
				accountStatus.hasAccount = true
				accountStatus.account = output.Account
			}

			operations.input = append(operations.input, accountOperation)
			operations.output = append(operations.output, output)
		case result["transaction"] != nil: // check json structure match transaction structure
			var transactionOperation TransactionOperation
			err := json.Unmarshal([]byte(line), &transactionOperation)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// convert string to time
			transactionOperation.Transaction.Time, _ = parseTime(transactionOperation.Transaction.Time.(string))

			// get output
			output := processTransaction(transactionOperation, accountStatus, operations.input)

			if output.Violations == nil {
				accountStatus.account.AvailableLimit = output.Account.AvailableLimit
			}

			operations.input = append(operations.input, transactionOperation)
			operations.output = append(operations.output, output)
		default:
			fmt.Println("operation not valid")
		}
	}

	return operations.output
}

func parseTime(data string) (time.Time, error) {
	loc, _ := time.LoadLocation("Etc/GMT")
	return time.ParseInLocation(time.RFC3339, data, loc)
}

func hasDoubledTransaction(slice []interface{}, transaction Transaction) bool {
	for i := len(slice) - 1; i >= 0; i-- {
		if reflect.TypeOf(slice[i]).String() == "main.TransactionOperation" {
			value := slice[i].(TransactionOperation).Transaction
			if value.Merchant == transaction.Merchant && value.Amount == transaction.Amount {
				if transaction.Time.(time.Time).Sub(value.Time.(time.Time)).Minutes() < 2 {
					return true
				}
			}
		}
	}
	return false
}

func getPivot(slice []interface{}, vType string) interface{} {
	for i := range slice {
		if reflect.TypeOf(slice[i]).String() == vType {
			return slice[i]
		}
	}
	return nil
}

func hasHighFrequencySmallInterval(slice []interface{}, transaction Transaction) bool {
	length := len(slice)
	if length > 3 {
		result := getPivot(slice, "main.AccountOperation")
		if result != nil {
			pivot := result.(AccountOperation).Account
			if pivot.ActiveCard && pivot.AvailableLimit == 100 {
				for count, i := 0, length-1; i >= 0; i-- {
					if reflect.TypeOf(slice[i]).String() == "main.TransactionOperation" {
						t1 := transaction.Time.(time.Time)
						t2 := slice[i].(TransactionOperation).Transaction.Time.(time.Time)
						if t1.Sub(t2).Minutes() < 2 {
							count++
							if count == 3 {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}
