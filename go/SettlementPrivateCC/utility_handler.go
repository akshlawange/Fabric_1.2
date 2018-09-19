package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//UtilityHandler provides APIs used to perform operations on CC's KV store
type utilityHandler struct {
}

// NewCollateralHandler create a new reference to CertHandler
func NewUtilityHandler() *utilityHandler {
	return &utilityHandler{}
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func (t *utilityHandler) getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Println("###### SettlementContract: function: getQueryResultForQueryString ")
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	if queryString == "" {
		return nil, errors.New("Incorrect number of arguments. Expecting queryString")
	}
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *utilityHandler) readSingleJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	Key := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			Key = append(Key, attributes[cnt])
		}
	}

	compositeKey, _ := stub.CreateCompositeKey(objectType, Key)
	JSONBytes, _ := stub.GetState(compositeKey)
	return JSONBytes, nil
}


func (t *utilityHandler) readMultiJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	keysIter, _ := stub.GetStateByPartialCompositeKey(objectType, partialKey)
	defer keysIter.Close()

	for keysIter.HasNext() {
		keyValue, _ := keysIter.Next()
		// Split keyvalue into key[String] aqnd value[byte[]]
		key := keyValue.Key
		value := keyValue.Value
		fmt.Println("JSONKey :[%s], JSONBytes[%s]", key, CToGoString(value))
		jsonString := CToGoString(value)
		jsonString = strconv.Quote(jsonString)
		if jsonString != "" {
			var settInst SettlementInstruction
			err := json.Unmarshal([]byte(value), &settInst)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}
			if settInst.ActiveInd == "A" {	
				finaldata = append(finaldata, value...)
				suffix := ","
				finaldata = append(finaldata, suffix...)
			}
		}
	}

	if len(finaldata) > 1 {
		finaldata = finaldata[:len(finaldata)-1]
		return finaldata, nil
	}

	fmt.Println("JSON :: [%s]", string(finaldata))
	return finaldata, nil
}

func (t *utilityHandler) readMultiSettlementInstructions(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	keysIter, _ := stub.GetStateByPartialCompositeKey(objectType, partialKey)
	defer keysIter.Close()
	prefix := "{\"SettlementInst\" : ["
	finaldata = append(finaldata, prefix...)
	for keysIter.HasNext() {
		keyValue, _ := keysIter.Next()
		// Split keyvalue into key[String] aqnd value[byte[]]
		key := keyValue.Key
		value := keyValue.Value
		fmt.Println("JSONKey :[%s], JSONBytes[%s]", key, CToGoString(value))
		jsonString := CToGoString(value)
		jsonString = strconv.Quote(jsonString)
		if jsonString != "" {
			var settInst SettlementInstruction
			err := json.Unmarshal([]byte(value), &settInst)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}
			if settInst.ActiveInd == "A" {
				finaldata = append(finaldata, value...)
				suffix := ","
				finaldata = append(finaldata, suffix...)
			}
		}
	}

	if len(finaldata) > 1 {
		finaldata = finaldata[:len(finaldata)-1]
		//return finaldata, nil
	}
	suffix := "]}"
	finaldata = append(finaldata, suffix...)

	fmt.Println("JSON :: [%s]", string(finaldata))
	return finaldata, nil
}

func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}
