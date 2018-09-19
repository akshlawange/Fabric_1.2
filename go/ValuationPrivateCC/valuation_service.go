package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type CollateralDetails struct {
	ObjectType                   string `json:"ObjectType,omitempty"`
	ProcessingSystemReference    string `json:"ProcessingSystemReference,omitempty"`
	CollateralType	             string `json:"CollateralType,omitempty"` //PLEDGE, BORROW, REPO
	LenderParticipantID          string `json:"LenderParticipantID,omitempty"`   //Lender of Collateral
	LenderParticipantAcct		 string `json:"LenderParticipantAcct,omitempty"`   //Lender of Collateral
	BorrowerParticipantID        string `json:"BorrowerParticipantID,omitempty"` //Borrower of Collareral
	BorrowerParticipantAcct      string `json:"BorrowerParticipantAcct,omitempty"` //Borrower of Collareral
	Instrument                   string `json:"Instrument,omitempty"`
	CusipDescription             string `json:"CusipDescription,omitempty"`
	AssetClass                   string `json:"AssetClass,omitempty"`
	SubAccount                   string `json:"SubAccount,omitempty"`
	TransactionDate              string `json:"TransactionDate,omitempty"`
	TransactionTimestamp         string `json:"TransactionTimestamp,omitempty"`
	EffectiveDate                string `json:"EffectiveDate,omitempty"`        //Same as TransactionDate
	ContractualValueDate         string `json:"ContractualValueDate,omitempty"`
	CloseEventDate               string `json:"CloseEventDate,omitempty"`
	Quantity                     string `json:"Quantity,omitempty"`
	CleanPrice                   string `json:"CleanPrice,omitempty"`
	DirtyPrice                   string `json:"DirtyPrice,omitempty"`
	Principal                    string `json:"Principal,omitempty"`
	Haircut                      string `json:"Haircut,omitempty"`
	AccruedInterestNoOfDays      int    `json:"AccruedInterestNoOfDays,string,omitempty"` // COUPON ACC DAYSS
	CouponAccruedInterest        string `json:"CouponAccruedInterest,omitempty"`
	Factor                       int    `json:"Factor,string,omitempty"`
	NetConsiderationBaseCurrency string `json:"NetConsiderationBaseCurrency,omitempty"`
	CurrentQuantity              string `json:"CurrentQuantity,omitempty"`
	CurrentPrice                 string `json:"CurrentPrice,omitempty"`
	CurrentValue                 string `json:"CurrentValue,omitempty"`
	LastUpdatedUser              string `json:"LastUpdatedUser,omitempty"`
	DateTime                     string `json:"DateTime,omitempty"`
	EditFlag                     string `json:"EditFlag,omitempty"`
	Version                      int    `json:"Version,string,omitempty"` // UPDATED BY CHAINCODE
	ActiveInd                    string `json:"ActiveInd,omitempty"`      // UPDATED BY CHAINCODE
}

var repHandler = NewRepoHandler()

type ValuationServiceChaincode struct {
}

func (t *ValuationServiceChaincode) initiateCollateralValuation(stub shim.ChaincodeStubInterface, participantID string, repoChaincodeID string, repoChannelID string, referenceDataChaincodeID string, globalChannelID string) pb.Response {

	if participantID != "" || repoChaincodeID != "" || repoChannelID != "" || referenceDataChaincodeID != "" || globalChannelID != "" {

		fmt.Println("###### ValuationCC: function: initiateCollateralValuation ")

		fmt.Printf("Query Repo CC : function initiateCollateralValuation for ", participantID)
		collList, err := repHandler.readRepoContracts(stub, participantID, repoChaincodeID, repoChannelID)
		if err != nil {
			return shim.Error("Failed to query collaterals for participantID")
		}
		fmt.Printf("Collateral for Participant 1. Got response %s", string(collList))

		collateralListJson, err := t.revaluateCollateralList(stub, string(collList), referenceDataChaincodeID, globalChannelID)
		if err != nil {
			return shim.Error("Failed to revaluate collaterals for participantID1")
		}

		return shim.Success([]byte(collateralListJson))
	}

	return shim.Error("Not enough parameters are passed!")

}

func (t *ValuationServiceChaincode) revaluateCollateralList(stub shim.ChaincodeStubInterface, collateralList string, referenceDataChaincodeID string, globalChannelID string) ([]byte, error) {

	fmt.Println("###### ValuationCC: function: revaluateCollateralList ")

	var arbitrary_json map[string]interface{}
	var err error
	var finaldata []byte

	fmt.Println("Repo Data JSON received: %v", collateralList)
	err = json.Unmarshal([]byte(collateralList), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		return nil, err
	}

	collateralData := arbitrary_json["Collateral"].([]interface{})
	fmt.Println("Collateral Data: %v", collateralData)

	for key1, value1 := range collateralData {
		fmt.Printf("Collateral Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
		jsonByte, err := json.Marshal(value1)
		jsonStr := convertArray(jsonByte)
		fmt.Println("Collateral Data JSON Str: %v", jsonStr)
		//Capture exchange rate
		if value1 != nil {
			var collStruct CollateralDetails
			err = json.Unmarshal([]byte(jsonStr), &collStruct)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
				return nil, err
			}

			if collStruct.Instrument != "" {
				priceByte, err := repHandler.readPriceByInstrument(stub, collStruct.Instrument, referenceDataChaincodeID, globalChannelID)
				if err != nil {
					fmt.Println("Failed to get prices for instrument ", collStruct.Instrument, err)
				}

				if string(priceByte) != "EMPTY" {
					price, _ := strconv.ParseFloat(string(priceByte), 64)
					currentQty, _ := strconv.ParseFloat(collStruct.CurrentQuantity, 64)
					newvalue := currentQty * price
					collStruct.CurrentValue = strconv.FormatFloat(newvalue, 'f', 2, 64)
					collStruct.CurrentPrice = strconv.FormatFloat(price, 'f', 2, 64)
					collStruct.LastUpdatedUser = "ValuationCC"

					jsonCollUpdate, err := json.Marshal(collStruct)
					if err != nil {
						fmt.Println("Failed to marshal collateral update json")

					}
					finaldata = append(finaldata, jsonCollUpdate...)
					suffix := ","
					finaldata = append(finaldata, suffix...)

					/*
						err = repHandler.updateRepoCollateral(stub, string(jsonCollUpdate), repoChaincodeID, repoChannelID)
						if err != nil {
							fmt.Println("Failed to update collateral with latest valuation ", collStruct.Instrument, err)
							//return err
						}
					*/
				}
			}
		}
	}

	if len(finaldata) > 1 {
		finaldata = finaldata[:len(finaldata)-1]
		var finaldata1 []byte
		prefix := "{\"Collateral\" : ["
		finaldata1 = append(finaldata1, prefix...)
		finaldata1 = append(finaldata1, finaldata...)
		suffix := "]}"
		finaldata1 = append(finaldata1, suffix...)
		fmt.Println("CollateralValuation FinalLIst::", finaldata1)

		return []byte(finaldata1), nil
	}

	return nil, nil
}

func (t *ValuationServiceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("###### ValuationCC: function: Invoke ")

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Function: %v %v", function, args)

	if function[0:1] == "i" {
		return t.invoke(stub, args[0], args) // old invoke function
	}
	if function[0:1] == "q" {
		return t.query(stub, args[0], args) // old query function
	}
	return shim.Error("Invoke: Invalid Function Name - function names begin with a q or i")
}

func (t *ValuationServiceChaincode) invoke(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {

	fmt.Println("###### ValuationCC: function: invoke ")
	fmt.Println("length JSON Data: %v %v", args[0], len(args))
	fmt.Println("[ValuationServiceChaincode] invoke args:", args[0], args[1], args[2], args[3], args[4])

	if function == "initiateCollateralValuation" {
		participantID := args[1]
		repoChaincodeID := args[2]
		repoChannelID := args[3]
		referenceDataChaincodeID := args[4]
		globalChannelID := args[5]

		return t.initiateCollateralValuation(stub, participantID, repoChaincodeID, repoChannelID, referenceDataChaincodeID, globalChannelID)
	}
	return shim.Error("Received unknown function invocation")
}

func (t *ValuationServiceChaincode) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### ValuationCC: function: query ")

	if function == "" || len(args) < 4 {
		return shim.Error("Not enough parameters passed")
	}

	fmt.Println("length JSON Data: %v %v", args[0], len(args))

	return shim.Error("Received unknown function query invocation with function")
}

func (t *ValuationServiceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("###### ValuationCC: function: Init ")

	_, args := stub.GetFunctionAndParameters()
	fmt.Println("[ValuationServiceChaincode] Init")
	if len(args) != 0 {
		return shim.Error("Init Incorrect number of arguments. Expecting 0")
	}

	return shim.Success(nil)
}

func convertArray(x []byte) string {
	jsonStr := strings.TrimLeft(string(x), "[")
	jsonStr = strings.TrimRight(string(jsonStr), "]")
	return jsonStr
}

func main() {
	fmt.Println("###### ValuationCC: function: main ")
	//	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(ValuationServiceChaincode))
	if err != nil {
		fmt.Println("Error starting ValuationServiceChaincode: %s", err)
	}

}
