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
	expected := `{"account":{"active-card":true,"available-limit":100},"violations":[]}` + "\n" + `{"account":{"active-card":true,"available-limit":50},"violations":[]}`
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

func TestProcessAccountWithViolationAccountAlreadyInitialized(t *testing.T) {
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
		Violations: []string{AccountAlreadyInitialized},
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
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: "2019-02-13T10:00:00.000Z"}})
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Habbib's", Amount: 15, Time: "2019-02-13T11:00:00.000Z"}})
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 30, Time: "2019-02-13T12:00:00.000Z"}})

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 35},
		hasAccount: true,
	}

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   10,
		Time:     "2019-02-13T13:01:00.000Z",
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

func TestProcessTransactionWithViolationAccountNotInitialized(t *testing.T) {
	var operations []interface{}

	status := AccountStatus{
		account:    Account{ActiveCard: false, AvailableLimit: 0},
		hasAccount: false,
	}

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   10,
		Time:     "2019-02-13T13:01:00.000Z",
	}}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: false, AvailableLimit: 0},
		Violations: []string{AccountNotInitialized},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolationInsufficientLimit(t *testing.T) {
	var operations []interface{}
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 130, Time: "2019-02-13T12:00:00.000Z"}})

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 100},
		hasAccount: true,
	}

	loc, _ := time.LoadLocation("Etc/GMT")
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T13:00:00.000Z", loc)

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   130,
		Time:     t1,
	}}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: true, AvailableLimit: 100},
		Violations: []string{InsufficientLimit},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolationCardNotActive(t *testing.T) {
	var operations []interface{}
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: false, AvailableLimit: 100}})

	newOperation := TransactionOperation{Transaction{
		Merchant: "Test",
		Amount:   35,
		Time:     "2019-02-13T13:00:00.000Z",
	}}

	status := AccountStatus{
		account:    Account{ActiveCard: false, AvailableLimit: 100},
		hasAccount: true,
	}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: false, AvailableLimit: 100},
		Violations: []string{CardNotActive},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolationDoubledTransaction(t *testing.T) {
	var operations []interface{}
	loc, _ := time.LoadLocation("Etc/GMT")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T12:00:00.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T12:00:10.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T12:01:00.000Z", loc)
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
		Violations: []string{DoubledTransaction},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithViolationHighFrequencySmallInterval(t *testing.T) {
	var operations []interface{}
	loc, _ := time.LoadLocation("Etc/GMT")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:00.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Habbib's", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 20, Time: t3}})
	t4, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:31.000Z", loc)
	newOperation := TransactionOperation{Transaction{
		Merchant: "Subway",
		Amount:   20,
		Time:     t4,
	}}

	status := AccountStatus{
		account:    Account{ActiveCard: true, AvailableLimit: 40},
		hasAccount: true,
	}

	expected := AccountOperationOutput{
		Account:    Account{ActiveCard: true, AvailableLimit: 40},
		Violations: []string{HighFrequencySmallInterval},
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
	loc, _ := time.LoadLocation("Etc/GMT")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:00.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t3}})
	t4, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:08.000Z", loc)
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
		Violations: []string{DoubledTransaction, HighFrequencySmallInterval},
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
	loc, _ := time.LoadLocation("Etc/GMT")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:00.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t3}})
	t4, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:08.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t4}})
	t5, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:18.000Z", loc)
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
		Violations: []string{InsufficientLimit, HighFrequencySmallInterval},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithMultipleViolations_3(t *testing.T) {
	var operations []interface{}
	loc, _ := time.LoadLocation("Etc/GMT")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:00.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t3}})
	t4, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:08.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t4}})
	t5, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:18.000Z", loc)
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
		Violations: []string{InsufficientLimit, HighFrequencySmallInterval},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}

func TestProcessTransactionWithMultipleViolations_4(t *testing.T) {
	var operations []interface{}
	loc, _ := time.LoadLocation("Etc/GMT")
	operations = append(operations, AccountOperation{Account: Account{ActiveCard: true, AvailableLimit: 100}})
	t1, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:00.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "McDonald's", Amount: 10, Time: t1}})
	t2, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 20, Time: t2}})
	t3, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:01:01.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t3}})
	t4, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:08.000Z", loc)
	operations = append(operations, TransactionOperation{Transaction: Transaction{Merchant: "Burger King", Amount: 5, Time: t4}})
	t5, _ := time.ParseInLocation(time.RFC3339, "2019-02-13T11:00:18.000Z", loc)
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
		Violations: []string{InsufficientLimit, HighFrequencySmallInterval},
	}
	result := processTransaction(newOperation, status, operations)

	if reflect.DeepEqual(expected, result) {
		t.Logf("processTransaction(...) PASSED \nexpected: %v \nresult: %v", expected, result)
	} else {
		t.Errorf("processTransaction(...) FAILED \nexpected: %v \nresult: %v", expected, result)
	}
}
