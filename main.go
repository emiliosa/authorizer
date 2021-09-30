package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

type JSONTime struct {
	data string
}

func (value *JSONTime) parseJSONTime(time.Time, error) {
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	time.ParseInLocation(time.RFC3339, value.data, loc)
}

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
	ACCOUNT_NOT_INITIALIZED       = "account-not-initialized"
	ACCOUNT_ALREADY_INITIALIZED   = "account-already-initialized"
	CARD_NOT_ACTIVE               = "card-not-active"
	DOUBLED_TRANSACTION           = "doubled-transaction"
	INSUFFICIENT_LIMIT            = "insufficient-limit"
	HIGH_FRECUENCY_SMALL_INTERVAL = "high-frequency-small-interval"
)

func main() {
	operations := handle()
	out := output(operations)
	fmt.Println(out)
}

func output(slice []AccountOperationOutput) string {
	var data []string
	for i := range slice {
		jsonData, _ := json.Marshal(slice[i])
		data = append(data, string(jsonData))
	}
	return strings.Join(data[:], "\n")
}

func processAccount(operation AccountOperation, accountStatus AccountStatus) AccountOperationOutput {
	var violations []string
	var activeCard bool
	var availableLimit int

	if accountStatus.hasAccount {
		violations = []string{ACCOUNT_ALREADY_INITIALIZED}
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
		violations = []string{ACCOUNT_NOT_INITIALIZED}
	} else {
		account.ActiveCard = status.account.ActiveCard
		account.AvailableLimit = status.account.AvailableLimit - new.Transaction.Amount

		if account.AvailableLimit < 0 {
			violations = []string{INSUFFICIENT_LIMIT}
		}

		if !status.account.ActiveCard {
			violations = append(violations, CARD_NOT_ACTIVE)
		}

		if hasDoubledTransaction(operations, new.Transaction) {
			violations = append(violations, DOUBLED_TRANSACTION)
		}

		if hasHighFrequencySmallInterval(operations, new.Transaction) {
			violations = append(violations, HIGH_FRECUENCY_SMALL_INTERVAL)
		}

		if violations != nil {
			account.AvailableLimit = status.account.AvailableLimit
		}
	}

	return AccountOperationOutput{account, violations}
}

func handle() []AccountOperationOutput {
	var operations = Operations{}
	var accountStatus = AccountStatus{
		account:    Account{},
		hasAccount: false,
	}

	// Open our jsonFile
	file, err := os.Open("operations.txt")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our file so that we can parse it later on
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	//scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// set interface type for unstructured json
		var result map[string]interface{}
		json.Unmarshal([]byte(line), &result)

		switch true {
		case result["account"] != nil: // check json structure match account structure
			var accountOperation AccountOperation
			json.Unmarshal([]byte(line), &accountOperation)

			output := processAccount(accountOperation, accountStatus)

			if output.Violations == nil {
				accountStatus.hasAccount = true
				accountStatus.account = output.Account
			}

			operations.input = append(operations.input, accountOperation)
			operations.output = append(operations.output, output)
		case result["transaction"] != nil: // check json structure match transaction structure
			var transactionOperation TransactionOperation
			json.Unmarshal([]byte(line), &transactionOperation)

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
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
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
