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

// NewUtilityHandler create a new reference to UtilHandler
func NewUtilityHandler() *utilityHandler {
	return &utilityHandler{}
}

collection := "RepoDealCollection"

func (t *utilityHandler) readSingleJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readSingleJSON ")
	Key := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			Key = append(Key, attributes[cnt])
		}
	}

	compositeKey, _ := stub.CreateCompositeKey(objectType, Key)
	JSONBytes, _ := stub.GetPrivateData(collection,compositeKey)

	fmt.Println("Retrived JSON :: [%s]", string(JSONBytes))

	return JSONBytes, nil
}

func (t *utilityHandler) readMultiJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readMultiJSON ")
	//notiID := String(notificationID)
	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	keysIter, _ := stub.GetStateByPartialCompositeKey(objectType, partialKey)
	defer keysIter.Close()

	//prefix := ""
	//finaldata = append(finaldata, prefix...)
	test := "false"
	//var notStruct []notificationsDetails
	for keysIter.HasNext() {
		keyValue, _ := keysIter.Next()
		// Split keyvalue into key[String] aqnd value[byte[]]
		key := keyValue.Key
		value := keyValue.Value
		fmt.Println("JSONKey :[%s], JSONBytes[%s]", key, CToGoString(value))
		jsonString := CToGoString(value)
		jsonString = strconv.Quote(jsonString)
		if jsonString != "" {
			if test == "true" {
				prefix := ","
				finaldata = append(finaldata, prefix...)
			}
			if test == "false" {
				test = "true"
			}
			finaldata = append(finaldata, jsonString...)
		}
	}

	//prefix = "]"
	//finaldata = append(finaldata, prefix...)
	fmt.Println("JSON :: [%s]", string(finaldata))
	return finaldata, nil
}

func (t *utilityHandler) readMultiCollateralJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readMultiCollateralJSON ")
	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	fmt.Println("Partial Key", partialKey)
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
			var exCollStruct CollateralDetails
			err := json.Unmarshal([]byte(value), &exCollStruct)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}

			//version_int, _ := strconv.Atoi(version)
			if exCollStruct.ActiveInd == "A" {

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
	return nil, nil
}

func (t *utilityHandler) readMultiCollateralJSONByParticipant(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readMultiCollateralJSONByParticipant ")
	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	fmt.Println("Partial Key", partialKey)
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
			var exCollStruct CollateralDetails
			err := json.Unmarshal([]byte(value), &exCollStruct)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}

			if exCollStruct.ActiveInd == "A" {
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

	return nil, nil
}

func (t *utilityHandler) readMultiParticipantsJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readMultiParticipantsJSON ")
	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	fmt.Println("Partial Key", partialKey)
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
			var partyStruct ParticipantDetails
			err := json.Unmarshal([]byte(value), &partyStruct)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}

			if partyStruct.ActiveInd == "A" {
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
	return nil, nil
}

func (t *utilityHandler) readMultiTradeJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readMultiTradeJSON ")
	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	fmt.Println("Partial Key", partialKey)
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
			var tradeStruct TradeDetails
			err := json.Unmarshal([]byte(value), &tradeStruct)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}
			finaldata = append(finaldata, value...)
			suffix := ","
			finaldata = append(finaldata, suffix...)

		}
	}
	if len(finaldata) > 1 {
		finaldata = finaldata[:len(finaldata)-1]
		return finaldata, nil
	}
	return nil, nil
}

func (t *utilityHandler) readMultiTradeActiveJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readMultiTradeJSON ")
	var finaldata []byte
	partialKey := []string{}

	for cnt := 0; cnt < len(attributes); cnt++ {
		if attributes[cnt] != "" {
			partialKey = append(partialKey, attributes[cnt])
		}
	}

	fmt.Println("Partial Key", partialKey)
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
			var tradeStruct TradeDetails
			err := json.Unmarshal([]byte(value), &tradeStruct)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}

			if tradeStruct.ActiveInd == "A" {
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
	return nil, nil
}

func (t *utilityHandler) readTradeJSON(stub shim.ChaincodeStubInterface, objectType string, attributes []string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: readTradeJSON ")
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
			var exTradeStruct TradeDetails
			err := json.Unmarshal([]byte(value), &exTradeStruct)
			if err != nil {
				return nil, errors.New("Error unmarshalling structure")
			}
			if exTradeStruct.ActiveInd == "A" {
				return value, nil
			}
		}
	}
	return nil, nil
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

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func (t *utilityHandler) getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: getQueryResultForQueryString ")
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	if queryString == "" {
		return nil, errors.New("Incorrect number of arguments. Expecting queryString")
	}
	resultsIterator, err := stub.GetPrivateDataQueryResult(collection,queryString)
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

func String(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}
