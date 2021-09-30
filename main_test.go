package main

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestEmptyOutput(t *testing.T) {
	expected := ""
	result := output(nil)
	expectedOut, _ := json.Marshal(expected)
	resultOut, _ := json.Marshal(result)

	if reflect.DeepEqual(expected, result) {
		t.Logf("output(nil) PASSED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	} else {
		t.Errorf("output(nil) FAILED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	}
}

func TestNotEmptyOutput(t *testing.T) {
	in := []AccountOperationOutput{
		{Account: Account{
			ActiveCard:     true,
			AvailableLimit: 100,
		}, Violations: nil},
		{Account: Account{
			ActiveCard:     true,
			AvailableLimit: 50,
		}, Violations: nil},
	}
	expected := `{"account":{"active-card":true,"available-limit":100},"violations":null}` + "\n" + `{"account":{"active-card":true,"available-limit":50},"violations":null}`
	result := output(in)
	expectedOut, _ := json.Marshal(expected)
	resultOut, _ := json.Marshal(result)

	if reflect.DeepEqual(expected, result) {
		t.Logf("output(...) PASSED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	} else {
		t.Errorf("output(...) FAILED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	}
}

func TestProcessAccountWithoutViolations(t *testing.T) {
	account := Account{
		ActiveCard:     true,
		AvailableLimit: 100,
	}
	operation := AccountOperation{account}
	accountStatus := AccountStatus{
		account:    Account{},
		hasAccount: false,
	}
	expected := AccountOperationOutput{
		Account:    account,
		Violations: nil,
	}
	result := processAccount(operation, accountStatus)
	expectedOut, _ := json.Marshal(expected)
	resultOut, _ := json.Marshal(result)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processAccount(...) PASSED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	} else {
		t.Errorf("processAccount(...) FAILED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	}
}

func TestProcessAccountWithViolation_ACCOUNT_ALREADY_INITIALIZED(t *testing.T) {
	account := Account{
		ActiveCard:     true,
		AvailableLimit: 100,
	}
	operation := AccountOperation{account}
	accountStatus := AccountStatus{
		account:    account,
		hasAccount: true,
	}
	expected := AccountOperationOutput{
		Account:    account,
		Violations: []string{ACCOUNT_ALREADY_INITIALIZED},
	}
	result := processAccount(operation, accountStatus)
	expectedOut, _ := json.Marshal(expected)
	resultOut, _ := json.Marshal(result)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processAccount(...) PASSED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	} else {
		t.Errorf("processAccount(...) FAILED \nexpected: %v \nresult: %v", string(expectedOut), string(resultOut))
	}
}

func TestProcessTransactionWithoutViolations(t *testing.T) {
	var operations []interface{}
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: false, AvailableLimit: 100}})
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: "2019-02-13T10:00:00+00:00"}})
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Habbib's", Amount: 15, Time: "2019-02-13T11:00:00+00:00"}})
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 30, Time: "2019-02-13T12:00:00+00:00"}})

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 35},
		hasAccount: true,
	}

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   10,
		Time:     "2019-02-13T13:01:00+00:00",
	}}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: true, AvailableLimit: 25},
		Violations: nil,
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolation_ACCOUNT_NOT_INITIALIZED(t *testing.T) {
	var operations []interface{}

	status := AccountStatus{
		account:    Account{ActiveCard: false, AvailableLimit: 0},
		hasAccount: false,
	}

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   10,
		Time:     "2019-02-13T13:01:00+00:00",
	}}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: false, AvailableLimit: 0},
		Violations: []string{ACCOUNT_NOT_INITIALIZED},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolation_INSUFFICIENT_LIMIT(t *testing.T) {
	var operations []interface{}
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 130, Time: "2019-02-13T12:00:00+00:00"}})

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 100},
		hasAccount: true,
	}

	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T13:00:00+00:00", loc)

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   130,
		Time:     t1,
	}}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: true, AvailableLimit: 100},
		Violations: []string{INSUFFICIENT_LIMIT},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolation_CARD_NOT_ACTIVE(t *testing.T) {
	var operations []interface{}
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: false, AvailableLimit: 100}})

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   35,
		Time:     "2019-02-13T13:00:00+00:00",
	}}

	status := AccountStatus{
		account:    Account{ActiveCard: false, AvailableLimit: 100},
		hasAccount: true,
	}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: false, AvailableLimit: 100},
		Violations: []string{CARD_NOT_ACTIVE},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolation_DOUBLED_TRANSACTION(t *testing.T) {
	var operations []interface{}
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T12:00:00+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T12:00:10+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T12:01:00+00:00", loc)
	newOperation := TransactionOperation{Transaction{
		Merchant: "McDonald's",
		Amount:   10,
		Time:     t3,
	}}

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 100},
		hasAccount: true,
	}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: true, AvailableLimit: 100},
		Violations: []string{DOUBLED_TRANSACTION},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithMultipleViolations_1(t *testing.T) {
	var operations []interface{}
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:00+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:01+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:01+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t3}})
	t4, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:08+00:00", loc)
	newOperation := TransactionOperation{Transaction{
		Merchant: "Burger King",
		Amount:   5,
		Time:     t4,
	}}

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 65},
		hasAccount: true,
	}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: true, AvailableLimit: 65},
		Violations: []string{DOUBLED_TRANSACTION, HIGH_FRECUENCY_SMALL_INTERVAL},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithMultipleViolations_2(t *testing.T) {
	var operations []interface{}
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:00+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:01+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:01+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t3}})
	t4, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:08+00:00", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t4}})
	t5, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:18+00:00", loc)
	newOperation := TransactionOperation{Transaction{
		Merchant: "Burger King",
		Amount:   150,
		Time:     t5,
	}}

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 65},
		hasAccount: true,
	}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: true, AvailableLimit: 65},
		Violations: []string{INSUFFICIENT_LIMIT, HIGH_FRECUENCY_SMALL_INTERVAL},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}
