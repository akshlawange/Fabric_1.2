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

func (t *SettlementContract) captureSettlementInstruction(stub shim.ChaincodeStubInterface, collection string, settInstJSON string, assetTokenChaincoodeID string, assetTokenChannelID string) pb.Response {
	
	fmt.Println("###### SettlementContract: function: captureSettlementInstruction ")
	var err error
	var settInstStruct SettlementInstruction		
	
	err = json.Unmarshal([]byte(settInstJSON), &settInstStruct)
	if err != nil {
		fmt.Println("captureSettlementInstruction: Error parsing JSON: ", err)
	}
			
	err = repHandler.newSettlementInstruction(stub, collection, settInstStruct)
	if err != nil {
		fmt.Println("captureSettlementInstruction: Failed to add new instruction", err)
		return shim.Error("Failed to add new token")
	}	

	return shim.Success([]byte("SUCCESS"))
}

func (t *SettlementContract) validateSettInstruction(stub shim.ChaincodeStubInterface, settInstJSON string, assetTokenChaincoodeID string, assetTokenChannelID string) pb.Response {
	
	fmt.Println("###### SettlementContract: function: validateSettInstruction ")
	var err error
	var arbitrary_json map[string]interface{}
	var tokenPositions []byte
	
	fmt.Println("validateSettInstruction: %v", settInstJSON)
	err = json.Unmarshal([]byte(settInstJSON), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		return shim.Error("Error parsing JSON")
	}
	
	instructionData := arbitrary_json["AssetTransfers"].([]interface{})
	fmt.Println("Instruction Data: %v", instructionData)
	prefix := "{\"Position\" : ["
	tokenPositions = append(tokenPositions, prefix...)

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
			tokenPos, err := t.checkPositionAvailability(stub, instStruct, assetTokenChaincoodeID, assetTokenChannelID)
			if err != nil {
				fmt.Println("Error checking position availbility: ", err)				
			}			
			
			tokenPositions = append(tokenPositions, tokenPos...)
			suffix := ","
			tokenPositions = append(tokenPositions, suffix...)

			if string(tokenPositions) == "FALSE" {
				return shim.Success([]byte("FALSE"))
			}
		}
	}

	if len(tokenPositions) > 1 {
		tokenPositions = tokenPositions[:len(tokenPositions)-1]		
	}
	suffix := "]}"
	tokenPositions = append(tokenPositions, suffix...)

	return shim.Success([]byte(tokenPositions))
}


func (t *SettlementContract) initiateSettlement(stub shim.ChaincodeStubInterface, collection string, settInstJSON string, tokenPositionJSON string, assetTokenChaincoodeID string, assetTokenChannelID string) pb.Response {
	fmt.Println("###### SettlementContract: function: initiateSettlement ")
	var err error
	var arbitrary_json map[string]interface{}

	var settInstStruct SettlementInstruction		
	var settAckJSON []byte
	//CHECK IF ALL COLLATERAL POSITIONS ARE AVAIABLE FOR SETTLEMENT
	err = json.Unmarshal([]byte(settInstJSON), &settInstStruct)
	if err != nil {
		fmt.Println("initiateSettlement: Error parsing JSON: ", err)
	}			

	if string(settInstJSON) != "" {		
		fmt.Println("initiateSettlement:Settlement Instruction: %v", settInstJSON)
		err = json.Unmarshal([]byte(settInstJSON), &arbitrary_json)
		if err != nil {
			fmt.Println("Error parsing JSON: ", err)
			return shim.Error("Error parsing JSON")
		}
	
		instructionData := arbitrary_json["AssetTransfers"].([]interface{})
		fmt.Println("initiateSettlement:Instruction Data: %v", instructionData)
		for key1, value1 := range instructionData {
			fmt.Printf("initiateSettlement:Instruction Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
			jsonByte, err := json.Marshal(value1)
			fmt.Println("initiateSettlement:Instruction Data JSON Str: ", string(jsonByte))

			//FOR EACH INSTRUCTION
			if string(jsonByte) != "" {
				var instStruct AssetTransfers
				err = json.Unmarshal([]byte(jsonByte), &instStruct)
				if err != nil {
					fmt.Println("Error parsing JSON: ", err)
					return shim.Error("Error parsing JSON")
				}
				fmt.Println("initiateSettlement:Settlement position : ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct, string(instStruct.PositionQty) )
				var arbitrary_pos map[string]interface{}
				err = json.Unmarshal([]byte(tokenPositionJSON), &arbitrary_pos)
				if err != nil {
					fmt.Println("Error parsing JSON: ", err)
					return shim.Error("Error parsing JSON")
				} 

				//Settlement Qty on the instruction
				var remainingSettQtyFloat float64								
				remainingSettQtyFloat, _ = strconv.ParseFloat(instStruct.PositionQty, 64)
		
				positionData := arbitrary_pos["Position"].([]interface{})
				fmt.Println("initiateSettlement:Position Data: %v", positionData)				

				for pKey, pValue := range positionData {
					fmt.Printf("initiateSettlement:Position Data index:%s  value1:%v  kind:%s  type:%s\n", pKey, pValue, reflect.TypeOf(pValue).Kind(), reflect.TypeOf(pValue))
					posJsonByte, err := json.Marshal(pValue)
					fmt.Println("initiateSettlement:Position Data JSON Str: ", string(posJsonByte))

					if string(posJsonByte) != "" {												
						
						var assetToken AssetToken
						err = json.Unmarshal([]byte(posJsonByte), &assetToken)
						if err != nil {
							fmt.Println("Error parsing JSON: ", err)
							return shim.Error("Error parsing JSON")
						}

						if assetToken.InstrumentID == instStruct.InstrumentID && assetToken.OwnerParty == instStruct.FromParty && assetToken.OwnerAcct == instStruct.FromAcct {
							availableQtyFloat, _ := strconv.ParseFloat(assetToken.PositionQty, 64)	
							
							if availableQtyFloat == remainingSettQtyFloat  {								
								fmt.Println("initiateSettlement:checking for tokens with equal quantity")							

								fmt.Println("initiateSettlement:tokenTranfer::", string(posJsonByte))										
								err = tHandler.tokenTranfer(stub, string(posJsonByte), instStruct.FromParty, instStruct.ToParty, instStruct.ToAcct, assetToken.PositionQty, assetTokenChaincoodeID, assetTokenChannelID)
								if err != nil {
									fmt.Println("Error during token trasnfers: ", err)
									return shim.Error("Error parsing JSON")
								}
								remainingSettQtyFloat = remainingSettQtyFloat - availableQtyFloat
								
								if remainingSettQtyFloat == 0 {
									//COMPLETED SETTLEMENT REQUIREMENT
									fmt.Println("initiateSettlement:Tokens found for required settlement")			
									break
								}
							} else if availableQtyFloat < remainingSettQtyFloat {
									
								fmt.Println("initiateSettlement:checking for tokens with less quantity")							


								fmt.Println("initiateSettlement:tokenTranfer::", string(posJsonByte))										
								err = tHandler.tokenTranfer(stub, string(posJsonByte), instStruct.FromParty, instStruct.ToParty, instStruct.ToAcct, assetToken.PositionQty, assetTokenChaincoodeID, assetTokenChannelID)
								if err != nil {
									fmt.Println("Error during token trasnfers: ", err)
									return shim.Error("Error parsing JSON")
								}

								remainingSettQtyFloat = remainingSettQtyFloat - availableQtyFloat

								if remainingSettQtyFloat == 0 {
									//COMPLETED SETTLEMENT REQUIREMENT
									fmt.Println("initiateSettlement:Tokens found for required settlement")			
									break
								}
								
							} else if availableQtyFloat > remainingSettQtyFloat {
								
								fmt.Println("initiateSettlement:checking for tokens with greater quantity")							
								fmt.Println("initiateSettlement:tokenTranfer::", string(posJsonByte))									
								remainingSettQtyStr := strconv.FormatFloat(remainingSettQtyFloat, 'f', 2, 64)
								
								err = tHandler.tokenTranfer(stub, string(posJsonByte), instStruct.FromParty, instStruct.ToParty, instStruct.ToAcct, remainingSettQtyStr, assetTokenChaincoodeID, assetTokenChannelID)
								if err != nil {
									fmt.Println("Error during token trasnfers: ", err)
									return shim.Error("Error parsing JSON")
								}

								remainingSettQtyFloat = remainingSettQtyFloat - remainingSettQtyFloat

								if remainingSettQtyFloat == 0 {
									//COMPLETED SETTLEMENT REQUIREMENT
									fmt.Println("initiateSettlement:Tokens found for required settlement")			
									break
								}							
							}	
						}
					}
				}
			}
		}

		settInstStruct.SettlementStatus = "SETTLED"
		settInstStruct.Reason = "SETTLED"		
		//UPDATE SETTLEMENT INSTRUCTION
		err = repHandler.updateSettlementInstruction(stub, collection, settInstStruct)
		if err != nil {
			fmt.Println("initiateSettlement: Failed to update settlement instruction: ", err)								
		}
		
		// CHECK LATEST REPO STATUS AND GENERATE EVENT FOR REPO ACK
		settAckJSON, err = repohandler.repoSettlementAcknowledge(stub, settInstStruct.SysReference, settInstStruct.SettlementType, settInstStruct.RepoChaincodeID, settInstStruct.RepoChannelID )
		if err != nil {
			fmt.Println("initiateSettlement: failed to update repo status", err)			
		}		
		
		fmt.Println("initiateSettlement: Sending an Settlement ACK event")					
		err = stub.SetEvent("RepoDealSettlementAck", []byte(settAckJSON))
		if err != nil {
			fmt.Println("SetEvent RepoDealSettlementAck Error", err)
		}

		return shim.Success([]byte("SUCCESS"))

	}
	return shim.Success([]byte("FAIL"))
}

func (t *SettlementContract) checkPositionAvailability(stub shim.ChaincodeStubInterface, instStruct AssetTransfers, assetTokenChaincoodeID string, assetTokenChannelID string) ([]byte, error) {
	
	var err error
	var remainingSettQtyFloat float64
	var tokenPositions []byte
	suffix := ","

	fmt.Println("###### SettlementContract: function: checkPositionAvailability ")
		
	availAsset, err := tHandler.readAvailableTokenPosition(stub, instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct, assetTokenChaincoodeID, assetTokenChannelID )
	if err != nil {
		fmt.Println("Failed to query for position ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct )				
	}
	fmt.Println("checkPositionAvailability: Settlement position : ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct, instStruct.PositionQty )
	fmt.Println("checkPositionAvailability: Available Position : ", string(availAsset))
	settQtyFloat, _ := strconv.ParseFloat(instStruct.PositionQty, 64)
	
	var arbitrary_pos map[string]interface{}
	err = json.Unmarshal([]byte(availAsset), &arbitrary_pos)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		return []byte("FALSE"), err
	}

	positionData := arbitrary_pos["Position"].([]interface{})
	fmt.Println("checkPositionAvailability:Position Data: %v", positionData)
	
	remainingSettQtyFloat = settQtyFloat

	for pKey, pValue := range positionData {
		fmt.Printf("checkPositionAvailability:Position Data index:%s  value1:%v  kind:%s  type:%s\n", pKey, pValue, reflect.TypeOf(pValue).Kind(), reflect.TypeOf(pValue))
		posJsonByte, err := json.Marshal(pValue)
		fmt.Println("checkPositionAvailability:Position Data JSON Str: ", string(posJsonByte))

		if string(posJsonByte) != "" {												
			
			var assetToken AssetToken
			err = json.Unmarshal([]byte(posJsonByte), &assetToken)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
				return []byte("FALSE"), err
			}
										
			availableQtyFloat, _ := strconv.ParseFloat(assetToken.PositionQty, 64)	
			fmt.Println("checkPositionAvailability:AvailableQty, SettlementQty ", availableQtyFloat, settQtyFloat)
			
			if availableQtyFloat == remainingSettQtyFloat  {								
				fmt.Println("checkPositionAvailability:checking for tokens with equal quantity")							
				fmt.Println("checkPositionAvailability:tokenTranfer::", string(posJsonByte))										
				
				remainingSettQtyFloat = remainingSettQtyFloat - availableQtyFloat

				tokenPositions = append(tokenPositions, posJsonByte...)
				tokenPositions = append(tokenPositions, suffix...)

				if remainingSettQtyFloat == 0 {
					//COMPLETED SETTLEMENT REQUIREMENT

					fmt.Println("checkPositionAvailability:Tokens found for required settlement")			
					return []byte(tokenPositions), nil		
				}

			} else if availableQtyFloat < remainingSettQtyFloat {
					
				fmt.Println("checkPositionAvailability:checking for tokens with less quantity")							


				fmt.Println("checkPositionAvailability:tokenTranfer::", string(posJsonByte))										
				remainingSettQtyFloat = remainingSettQtyFloat - availableQtyFloat

				tokenPositions = append(tokenPositions, posJsonByte...)				
				tokenPositions = append(tokenPositions, suffix...)

				if remainingSettQtyFloat == 0 {				
					//COMPLETED SETTLEMENT REQUIREMENT
					
					fmt.Println("checkPositionAvailability:Tokens found for required settlement")			
					return []byte(tokenPositions), nil		
				}
				
			} else if availableQtyFloat > remainingSettQtyFloat {
				
				fmt.Println("checkPositionAvailability:checking for tokens with greater quantity")							
				fmt.Println("checkPositionAvailability:tokenTranfer::", string(posJsonByte))										

				remainingSettQtyFloat = remainingSettQtyFloat - remainingSettQtyFloat

				tokenPositions = append(tokenPositions, posJsonByte...)				
				tokenPositions = append(tokenPositions, suffix...)

				if remainingSettQtyFloat == 0 {
					//COMPLETED SETTLEMENT REQUIREMENT
				
					fmt.Println("checkPositionAvailability:Tokens found for required settlement")			
					return []byte(tokenPositions), nil		
				}			
			}
		}
	}

	fmt.Println("checkPositionAvailability:Not enough position available for : ", instStruct.InstrumentID, instStruct.FromParty, instStruct.FromAcct )			

	return []byte("FALSE"), nil		
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
		collection := args[1]
		jsonData := args[2]
		assetTokenCC := args[3]
		assetTokenChannelID := args[4]
		return t.captureSettlementInstruction(stub, collection, jsonData, assetTokenCC, assetTokenChannelID)

	} else if function == "validateSettInstruction" {
		jsonData := args[1]
		assetTokenCC := args[2]
		assetTokenChannelID := args[3]
		return t.validateSettInstruction(stub, jsonData, assetTokenCC, assetTokenChannelID)

	} else if function == "initiateSettlement" {
		collection := args[1]
		jsonData := args[2]
		positionJSONData := args[3]
		assetTokenCC := args[4]
		assetTokenChannelID := args[5]
		return t.initiateSettlement(stub, collection, jsonData, positionJSONData, assetTokenCC, assetTokenChannelID)
	} 
	return shim.Error("ERROR: Received unknown function invocation")
}

func (t *SettlementContract) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### SettlementContract: function: query ")

	fmt.Println("[SettlementContract] Query")
	
	if function == "getSettInstInformation" {
		collection := args[1]
		sysReference := args[2]
		tradeType := args[3]
		assetType := args[4]
		settlementType := args[5]

		JSONbytes, err := t.getSettInstInformation(stub, collection, sysReference, tradeType, assetType, settlementType )
		if err != nil {
			return shim.Error("Error querying token")
		}
		return shim.Success(JSONbytes)

	} else if function == "getSettInstByFilter" {
		collection := args[1]
		queryString := args[2]
		return t.getSettInstByFilter(stub, collection, queryString)
	} 
	return shim.Error("SettlementContract: Received unknown function query invocation with function ")
}

func (t *SettlementContract) getSettInstByFilter(stub shim.ChaincodeStubInterface, collection string, queryString string) pb.Response {
	
	fmt.Println("###### SettlementContract: function:getSettInstByFilter ", collection, queryString)

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, collection, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}
	
func (t *SettlementContract) getSettInstInformation(stub shim.ChaincodeStubInterface, collection string, sysReference string, tradeType string, assetType string, settlementType string) ([]byte, error) {
	
	fmt.Println("###### SettlementContract: function:getSettInstInformation ")
	
	fmt.Println("getSettInstInformation: Querying for : [%s][%s][%s][%s][%s]", string(collection), string(sysReference), string(tradeType), string(assetType), string(settlementType) )
	
	assetToken, err := repHandler.querySettlementInstruction(stub, collection, sysReference, tradeType, assetType, settlementType)
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
	

