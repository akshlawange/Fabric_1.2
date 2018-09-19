package main

import (
	"encoding/json"
	"fmt"
	"reflect"		
	"strconv"
	"strings"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SettlementContract struct {

}

var repHandler = NewRepositoryHandler()
var tHandler = NewTokenHandler()
var repohandler = NewRepoHandler()


func (t *SettlementContract) captureSettlementInstruction(stub shim.ChaincodeStubInterface, settInstJSON string) pb.Response {
	
	fmt.Println("###### SettlementContract: function: captureSettlementInstruction ")
	var err error
	var settInstStruct SettlementInstruction		
	
	err = json.Unmarshal([]byte(settInstJSON), &settInstStruct)
	if err != nil {
		fmt.Println("captureSettlementInstruction: Error parsing JSON: ", err)
	}
			
	err = repHandler.newSettlementInstruction(stub, settInstStruct)
	if err != nil {
	fmt.Println("captureSettlementInstruction: Failed to add new instruction", err)
		shim.Error("Failed to add new token")
	}	
	
	return shim.Success([]byte("SUCCESS"))
}


func (t *SettlementContract) initiateSettlement(stub shim.ChaincodeStubInterface, settInstJSON string, assetTokenChaincoodeID string, assetTokenChannelID string) pb.Response {
	
	fmt.Println("###### SettlementContract: function: initiateSettlement ")
	var err error
	var settInstStruct SettlementInstruction		
	//var goForSett string
	//CHECK IF ALL COLLATERAL POSITIONS ARE AVAIABLE FOR SETTLEMENT

	goForSett, err := t.validateSettInstruction(stub, settInstJSON, assetTokenChaincoodeID, assetTokenChannelID )

	if string(goForSett) == "FALSE" {
		
		return shim.Success([]byte("NOTAVAILABLE"))	

	} else {
		
		err = t.transferAssets(stub, settInstJSON, assetTokenChaincoodeID, assetTokenChannelID)
		if err != nil {
			fmt.Println("initiateSettlement: Failed to transfer: ", err)								
			settInstStruct.SettlementStatus = "FAILED"
			settInstStruct.Reason = "FAILED SETTLEMENT"	
			//UPDATE SETTLEMENT INSTRUCTION
			err = repHandler.updateSettlementInstruction(stub, settInstStruct)
			if err != nil {
				fmt.Println("initiateSettlement: Failed to update settlement instruction: ", err)								
			}		
			return shim.Success([]byte("FAILED"))	
		}

	}

	settInstStruct.SettlementStatus = "SETTLED"
	settInstStruct.Reason = "SETTLED"		
	//UPDATE SETTLEMENT INSTRUCTION
	err = repHandler.updateSettlementInstruction(stub, settInstStruct)
	if err != nil {
		fmt.Println("initiateSettlement: Failed to update settlement instruction: ", err)								
	}
	
	// CHECK LATEST REPO STATUS AND GENERATE EVENT FOR REPO ACK
	err = repohandler.repoSettlementAcknowledge(stub, settInstStruct.SysReference, settInstStruct.SettlementType, settInstStruct.RepoChaincodeID, settInstStruct.RepoChannelID )
	if err != nil {
		fmt.Println("initiateSettlement: failed to update repo status", err)			
	}
	
	fmt.Println("initiateSettlement: Settlement successful: ", err)			
	return shim.Success([]byte("SUCCESS"))
}

func (t *SettlementContract) validateSettInstruction(stub shim.ChaincodeStubInterface, settInstJSON string, assetTokenChaincoodeID string, assetTokenChannelID string) ([]byte, error) {
	
	fmt.Println("###### SettlementContract: function: validateSettInstruction ")
	var err error
	var arbitrary_json map[string]interface{}
	var noGoSettFlag string
	noGoSettFlag = "FALSE"
	
	fmt.Println("validateSettInstruction: %v", settInstJSON)
	err = json.Unmarshal([]byte(settInstJSON), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}
	
	instructionData := arbitrary_json["AssetTransfers"].([]interface{})
	fmt.Println("Instruction Data: %v", instructionData)

	for key1, value1 := range instructionData {
		fmt.Printf("Instruction Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
		jsonByte, err := json.Marshal(value1)
		//jsonStr = convertArray(jsonByte)
		fmt.Println("Instruction Data JSON Str: %v", string(jsonByte))

		if string(jsonByte) != "" {
			var instStruct AssetTransfers
			err = json.Unmarshal([]byte(jsonByte), &instStruct)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}

			//CHECK POSITION AVAILABILITY
			flag, err := t.checkPositionAvailability(stub, instStruct, assetTokenChaincoodeID, assetTokenChannelID)
			if err != nil {
				fmt.Println("Error checking position availbility: ", err)				
			}			
			if string(flag) == "FALSE" {
				noGoSettFlag = "TRUE"
			}			
		}
	}
	return []byte(noGoSettFlag), nil
}

func (t *SettlementContract) checkPositionAvailability(stub shim.ChaincodeStubInterface, instStruct AssetTransfers, assetTokenChaincoodeID string, assetTokenChannelID string) ([]byte, error) {
	
	var err error
	fmt.Println("###### SettlementContract: function: checkPositionAvailability ")
		
	availAsset, err := tHandler.readAvailableTokenPosition(stub, instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct, assetTokenChaincoodeID, assetTokenChannelID )
	if err != nil {
		fmt.Println("Failed to query for position ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct )				
	}
	fmt.Println("checkPositionAvailability: Settlement position : ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct )
	fmt.Println("checkPositionAvailability: Available Position : ", string(availAsset))
	
	var assetToken AssetToken
	err = json.Unmarshal([]byte(availAsset), &assetToken)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}
	
	availableQtyFloat, _ := strconv.ParseFloat(assetToken.PositionQty, 64)
	settQtyFloat, _ := strconv.ParseFloat(instStruct.PositionQty, 64)
		
	if availableQtyFloat < settQtyFloat {
		fmt.Println("checkPositionAvailability:Not enough position available for : ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct )		
		return []byte("FALSE"), nil
	}
	
	return []byte("TRUE"), nil
}

func (t *SettlementContract) transferAssets(stub shim.ChaincodeStubInterface, settInstJSON string,  assetTokenChaincoodeID string, assetTokenChannelID string) (error) {
	
	fmt.Println("###### SettlementContract: function: transferAssets ")
	var err error
	var arbitrary_json map[string]interface{}
	if string(settInstJSON) != "" {
		
		fmt.Println("transferAssets:Settlement Instruction: %v", settInstJSON)

		err = json.Unmarshal([]byte(settInstJSON), &arbitrary_json)
		if err != nil {
			fmt.Println("Error parsing JSON: ", err)
			return err
		}
	
		instructionData := arbitrary_json["AssetTransfers"].([]interface{})
		fmt.Println("transferAssets:Instruction Data: %v", instructionData)

		for key1, value1 := range instructionData {
			fmt.Printf("transferAssets:Instruction Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
			jsonByte, err := json.Marshal(value1)
			fmt.Println("transferAssets:Instruction Data JSON Str: ", string(jsonByte))

			if string(jsonByte) != "" {
				var instStruct AssetTransfers
				err = json.Unmarshal([]byte(jsonByte), &instStruct)
				if err != nil {
					fmt.Println("Error parsing JSON: ", err)
					return err
				}

				availAsset, err := tHandler.readAvailableTokenPosition(stub, instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct, assetTokenChaincoodeID, assetTokenChannelID )
				if err != nil {
					fmt.Println("Failed to query for position ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct )				
				}
				fmt.Println("transferAssets:Settlement position : ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct )
				fmt.Println("transferAssets:Available Position : ", string(availAsset))
		
				var assetToken AssetToken
				err = json.Unmarshal([]byte(availAsset), &assetToken)
				if err != nil {
					fmt.Println("Error parsing JSON: ", err)
					return err
				}
		
				availableQtyFloat, _ := strconv.ParseFloat(assetToken.PositionQty, 64)	
				settQtyFloat, _ := strconv.ParseFloat(instStruct.PositionQty, 64)
			
				if availableQtyFloat >= settQtyFloat {
				
					err = tHandler.tokenTranfer(stub, string(availAsset), instStruct.ToParty, instStruct.ToAcct, instStruct.PositionQty, assetTokenChaincoodeID, assetTokenChannelID)
					if err != nil {
						fmt.Println("Error during token trasnfers: ", err)
						return err
					}
				}
			}
		}
	}
	return nil
}

func (t *SettlementContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("###### SettlementContract: function: Invoke ")

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("SettlementContract: Function: %v %v", function, args)

	if function[0:1] == "i" {
		return t.invoke(stub, args[0], args) // old invoke function
	}
	if function[0:1] == "q" {
		return t.query(stub, args[0], args) // old query function
	}
	return shim.Error("SettlementContract: Invoke: Invalid Function Name - function names begin with a q or i")
}

func (t *SettlementContract) invoke(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### SettlementContract: function: invoke ")

	fmt.Println("[SettlementContract] Invoke")
	if function == "captureSettlementInstruction" {
		jsonData := args[1]
		return t.captureSettlementInstruction(stub, jsonData)

	} else if function == "initiateSettlement" {
		jsonData := args[1]
		assetTokenCC := args[2]
		assetTokenChannelID := args[3]
		return t.initiateSettlement(stub, jsonData, assetTokenCC, assetTokenChannelID)
	} 
	return shim.Error("ERROR: Received unknown function invocation")
}

func (t *SettlementContract) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### SettlementContract: function: query ")

	fmt.Println("[SettlementContract] Query")
	
	if function == "getSettInstInformation" {
		sysReference := args[1]

		JSONbytes, err := t.getSettInstInformation(stub, sysReference)
		if err != nil {
			return shim.Error("Error querying token")
		}
		return shim.Success(JSONbytes)

	} else if function == "getSettInstByFilter" {
		return t.getSettInstByFilter(stub, args[1])
	} 
	return shim.Error("SettlementContract: Received unknown function query invocation with function ")
}

func (t *SettlementContract) getSettInstByFilter(stub shim.ChaincodeStubInterface, queryString string) pb.Response {
	
	fmt.Println("###### SettlementContract: function:getSettInstByFilter ", queryString)

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)

}
	
func (t *SettlementContract) getSettInstInformation(stub shim.ChaincodeStubInterface, sysReference string) ([]byte, error) {
	
	fmt.Println("###### SettlementContract: function:getSettInstInformation ")
	
	fmt.Println("getSettInstInformation: Querying for : [%s][%s][%s][%s][%s]", string(sysReference))
	
	assetToken, err := repHandler.querySettlementInstruction(stub, sysReference)
	if err != nil {
		fmt.Println("SettlementContract: Failed to query a token row [%s]", err)
	}
	
	fmt.Println("getSettInstInformation: Asset Token Information: [%v][%v]", assetToken)
	
	return assetToken, nil
}
	
func (t *SettlementContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("###### SettlementContract: function:Init ")
	_, args := stub.GetFunctionAndParameters()
	fmt.Println("[SettlementContract] Init")
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
	//	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(SettlementContract))
	if err != nil {
		fmt.Println("SettlementContract: Error starting SettlementContract: %s", err)
	}
}
	

