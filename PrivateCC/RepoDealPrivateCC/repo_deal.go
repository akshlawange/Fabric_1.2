package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type RepoDealChaincode struct {
}

var partHandler = NewParticipantsHandler()
var collHandler = NewCollateralHandler()
var tHandler = NewTradeHandler()
var mpHandler = NewMultipartyHandler()
var instHandler = NewSettInstHandler()
var repHandler = NewRepoHandler()
var collection = "RepoDealCollection"

type RepoDeal struct {
	//RepoStatusSummary -- RepoDealRef, RepoStatus, RepoDealChainCodeID, InventoryMangementChaincodeID, MultipartyChaincodeID
	ObjectType            string              `json:"ObjectType,omitempty"`
	Party                 ParticipantDetails  `json:"Party,omitempty"`
	Counterparty          ParticipantDetails  `json:"Counterparty,omitempty"`
	PartyCustodian        ParticipantDetails  `json:"PartyCustodian,omitempty"`
	CounterpartyCustodian ParticipantDetails  `json:"CPartyCustodian,omitempty"`
	Trade                 TradeDetails        `json:"Trade,omitempty"`
	Collaterals           []CollateralDetails `json:"Collateral,omitempty"`
}

type SystemToken struct {
	SystemReferenceId int32 `json:"SystemReferenceId,string,omitempty"`
}

//WebAPP or External invoke service to call this function
func (t *RepoDealChaincode) repoDealCapture(stub shim.ChaincodeStubInterface, repoJSON string, action string, mpChaincodeID string, repoChannelID string) pb.Response {
	var err error
	var ProcessingSysRef string
	
	//Getting Processing System Reference from recieved JSON
	var tradeStruct TradeDetails
        var arbitrary_json map[string]interface{}
	err = json.Unmarshal([]byte(repoJSON), &arbitrary_json)
        if err != nil {
                fmt.Println("Error parsing JSON: ", err)
        }
	
	//Capturing Trade Details
	jsonByte, err := json.Marshal(arbitrary_json["Trade"])
        jsonStr := convertArray(jsonByte)
	err = json.Unmarshal([]byte(jsonStr), &tradeStruct)
        if err != nil {
        	fmt.Println("Error parsing JSON: ", err)
        }
	ProcessingSysRef = tradeStruct.ProcessingSystemReference
 
	fmt.Println("###### RepoDealChaincode: function: repoDealCapture ")

	if action == "NEW" {
		err := repHandler.repoDealDeployment(stub, ProcessingSysRef, repoJSON, "NEW", "PENDAPPROVAL")        
		if err != nil {
			fmt.Println("Error deploying  new Repo deal", err)
			return shim.Error(err.Error())
		}
		fmt.Println("systemReference::",ProcessingSysRef)

		err = mpHandler.newRepoDealNotification(stub, repoJSON, "New", ProcessingSysRef, mpChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error generating notification for new Repo deal", err)
			return shim.Error(err.Error())
		}

	} else if action == "AMEND" {

		err = mpHandler.newRepoDealNotification(stub, repoJSON, "Amend", "", mpChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error generating notification for Repo deal Amend", err)
			return shim.Error(err.Error())
		}

	} else if action == "CANCEL" {

		err = mpHandler.newRepoDealNotification(stub, repoJSON, "Cancel", "", mpChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error generating notification for Repo deal Cancel", err)
			return shim.Error(err.Error())
		}

	}
	return shim.Success([]byte("Repo Deal is captured successfully!"))

}

//WebAPP or External invoke service to call this function
func (t *RepoDealChaincode) repoCollateralSubCapture(stub shim.ChaincodeStubInterface, sysReference string, repoJSON string, action string, mpChaincodeID string, repoChannelID string) pb.Response {
	var err error
	fmt.Println("###### RepoDealChaincode: function: repoCollateralSubCapture ")
	if action == "AMEND" {
		//ADD LOGIC TO CHECK REPO STATUS *****IF PENDAPPROVAL update the JSON and cancel old notification and send new one
		err = mpHandler.newCollateralSubNotification(stub, sysReference, repoJSON, "CollateralSubstitution", mpChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error generating notification for Repo deal Collateral Substitution", err)
			return shim.Error(err.Error())
		}
	}
	return shim.Success([]byte("Collateral Sub request is captured successfully!"))

}

func (t *RepoDealChaincode) initiateRepoDealCollateralSub(stub shim.ChaincodeStubInterface, transactionRef string, repoJSON string, repoChaincodeID string, repoChannelID string) pb.Response {
	
	var err error
	var settInstJSONArray []byte
	var settlementType string
	settlementType = "RepoCollateralSubstitution"
	fmt.Println("###### RepoDealChaincode: function: initiateRepoDealCollateralSub ")
	prefix := "{\"Instruction\" : ["
	settInstJSONArray = append(settInstJSONArray, prefix...)
		
	postingJSON, err := repHandler.repoDealCollateralSub(stub, transactionRef, repoJSON, settlementType, repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error during collateral sub", err)
		return shim.Error(err.Error())
	}
	settInstJSONArray = append(settInstJSONArray, postingJSON...)		
	suffix := "]}"
	settInstJSONArray = append(settInstJSONArray, suffix...)

	err = stub.SetEvent("RepoDealCollSubPosting", []byte(settInstJSONArray))
	if err != nil {
		fmt.Println("SetEvent RepoDealCollSubPosting Error", err)
	}

	return shim.Success([]byte("Collateral Sub is performed successfully!"))
}

//MultiPartyCC to call this function
func (t *RepoDealChaincode) repoDealApproval(stub shim.ChaincodeStubInterface, transactionRef string, repoJSON string, notificationType string, multiPartyAction string) pb.Response {
	var err error
	var transactionAction string
	var newRepoStatus string
	var currentRepoStatus string
	var deployFlag string
	var sourceSystem string
	deployFlag = "N"
	sourceSystem = "Approval Service"
	fmt.Println("###### RepoDealChaincode: function: repoDealApproval ")

	currentRepoStatus, err = repHandler.getRepoStatus(stub, transactionRef)
	if err != nil {
		fmt.Println("Error reading Repo Deal status", err)
		return shim.Error(err.Error())
	}

	if multiPartyAction == "Approved" {

		if notificationType == "New" {
			if currentRepoStatus == "PENDAPPROVAL" {
				newRepoStatus = "OPENLEGPEND"
				deployFlag = "N"

				err = repHandler.repoStatusUpdate(stub, transactionRef, newRepoStatus, sourceSystem)
				if err != nil {
					fmt.Println("Error updating Repo Deal status", err)
					return shim.Error(err.Error())
				}
			}
		} else if notificationType == "Amend" {
			if currentRepoStatus == "REJECTED" {
				newRepoStatus = "OPENLEGPEND"
				transactionAction = "AMEND"
				deployFlag = "Y"

			} else if currentRepoStatus == "PENDAPPROVAL" {
				newRepoStatus = "OPENLEGPEND"
				transactionAction = "AMEND"
				deployFlag = "Y"

			} else if currentRepoStatus == "OPENLEGSETTLED" {
				newRepoStatus = "OPENLEGSETTLED" //NO ECONOMICAL CHANGE FOR SETTLEMENT
				transactionAction = "AMEND"
				deployFlag = "Y"
			}
		} else if notificationType == "Cancel" {
			if currentRepoStatus == "REJECTED" {
				newRepoStatus = "PENDAPPROVAL"
				transactionAction = "CANCEL"
				deployFlag = "Y"
			} else if currentRepoStatus == "PENDAPPROVAL" {
				newRepoStatus = "CANCELLED"
				transactionAction = "CANCEL"
				deployFlag = "Y"
			} else if currentRepoStatus == "OPENLEGPEND" {
				newRepoStatus = "CANCELLED"
				transactionAction = "CANCEL"
				deployFlag = "Y"
			}
		} else if notificationType == "CollateralSubstitution" {
			if currentRepoStatus == "OPENLEGSETTLED" {
				newRepoStatus = "COLLSUBPEND"
				transactionAction = "AMEND"

				err = repHandler.repoStatusUpdate(stub, transactionRef, newRepoStatus, sourceSystem)
				if err != nil {
					fmt.Println("Error updating Repo Deal status", err)
					return shim.Error(err.Error())
				}

			}
		} else if notificationType == "InterestPayment" {
			if currentRepoStatus == "OPENLEGSETTLED" {
				newRepoStatus = "INTPAYMENTPEND"
				transactionAction = "AMEND"

				err = repHandler.repoStatusUpdate(stub, transactionRef, newRepoStatus, sourceSystem)
				if err != nil {
					fmt.Println("Error updating Repo Deal status", err)
					return shim.Error(err.Error())
				}
			}
		} else if notificationType == "CashAdjustment" {
			if currentRepoStatus == "OPENLEGSETTLED" {
				newRepoStatus = "CASHADJPEND"
				transactionAction = "AMEND"

				err = repHandler.repoStatusUpdate(stub, transactionRef, newRepoStatus, sourceSystem)
				if err != nil {
					fmt.Println("Error updating Repo Deal status", err)
					return shim.Error(err.Error())
				}
			}
		} else if notificationType == "Close" {
			if currentRepoStatus == "OPENLEGSETTLED" {
				newRepoStatus = "CLOSELEGPEND"
				transactionAction = "AMEND"

				err = repHandler.repoStatusUpdate(stub, transactionRef, newRepoStatus, sourceSystem)
				if err != nil {
					fmt.Println("Error updating Repo Deal status", err)
					return shim.Error(err.Error())
				}
			}
		}

	} else if multiPartyAction == "Rejected" {

		if notificationType == "New" {
			if currentRepoStatus == "PENDAPPROVAL" {
				newRepoStatus = "REJECTED"
				deployFlag = "N"
				err = repHandler.repoStatusUpdate(stub, transactionRef, newRepoStatus, sourceSystem)
			}
			if currentRepoStatus == "REJECTED" {
				newRepoStatus = "REJECTED"
				deployFlag = "N"
			}
		} else if notificationType == "Amend" {
			if currentRepoStatus == "PENDAPPROVAL" {
				newRepoStatus = "REJECTED"
				deployFlag = "N"
				err = repHandler.repoStatusUpdate(stub, transactionRef, newRepoStatus, sourceSystem)
			} else if currentRepoStatus == "REJECTED" {
				newRepoStatus = "REJECTED"
				deployFlag = "N"
			} else if currentRepoStatus == "OPENLEGSETTLED" {
				newRepoStatus = "OPENLEGSETTLED"
				deployFlag = "N"
			}
		} else if notificationType == "Cancel" {
			if currentRepoStatus == "PENDAPPROVAL" {
				newRepoStatus = "PENDAPPROVAL"
				deployFlag = "N"
			} else if currentRepoStatus == "REJECTED" {
				newRepoStatus = "REJECTED"
				deployFlag = "N"
			} /*else if currentRepoStatus == "OPENLEGSETTLED" {
				newRepoStatus = "OPENLEGSETTLED"
				deployFlag = "N"
			} */
		}
	}

	if deployFlag == "Y" && notificationType == "New" || notificationType == "Amend" || notificationType == "Cancel" {
		err := repHandler.repoDealDeployment(stub, transactionRef, repoJSON, transactionAction, newRepoStatus)
		if err != nil {
			fmt.Println("Error deploying  new Repo deal", err)
			return shim.Error(err.Error())
		}
		fmt.Println("Repo Deal is deployed successfully", transactionRef)

	}
	return shim.Success([]byte("Repo Deal is captured successfully!"))
}

func (t *RepoDealChaincode) initiateInterestPaymentRequest(stub shim.ChaincodeStubInterface, sysReference string, payment string, party string, counterparty string, tranType string, tranOriginatorParty string, mpChaincodeID string, repoChannelID string) pb.Response {
	fmt.Println("###### RepoDealChaincode: function: initiateInterestPaymentRequest ")

	var err error
	err = mpHandler.newInterestPaymentNotification(stub, sysReference, payment, party, counterparty, tranType, tranOriginatorParty, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error generating notification for interest payment", err)
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Repo Deal is captured successfully!"))
}

// inidicator == CREDIT OR DEBIT
func (t *RepoDealChaincode) initiateCashAdjustmentRequest(stub shim.ChaincodeStubInterface, sysReference string, payment string, indicator string, party string, counterparty string, tranType string, tranOriginatorParty string, mpChaincodeID string, repoChannelID string) pb.Response {
	fmt.Println("###### RepoDealChaincode: function: initiateCashAdjustmentRequest ")

	var err error
	err = mpHandler.newCashAdjustmentNotification(stub, sysReference, payment, indicator, party, counterparty, tranType, tranOriginatorParty, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error generating notification for cash adjustment", err)
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Cash Adjustment is captured successfully!"))
}

func (t *RepoDealChaincode) initiateAutoRepoDealClose(stub shim.ChaincodeStubInterface, sysReference string, repoChaincodeID string, repoChannelID string) pb.Response {
	fmt.Println("###### RepoDealChaincode: function: initiateAutoRepoDealClose ")

	var err error
	var settlementType string
	settlementType = "RepoCloseLegSettlement"
	var settInstJSONArray []byte
	prefix := "{\"Instruction\" : ["
	var delimiter string
	delimiter = ","

	stockJSON, cashJSON, err := repHandler.repoDealCloseSettlement(stub, sysReference, settlementType, repoChaincodeID, repoChannelID ) 
	if err != nil {
		fmt.Println("Error repoDealCloseSettlement", err)
		return shim.Error(err.Error())
	}

	settInstJSONArray = append(settInstJSONArray, prefix...)
	if cashJSON != nil {	
		fmt.Println("cashJSON :", string(cashJSON))	
		settInstJSONArray = append(settInstJSONArray, cashJSON...)		
		settInstJSONArray = append(settInstJSONArray, delimiter...)	
	}

	if stockJSON != nil {
		fmt.Println("stockJSON :", string(stockJSON))
		settInstJSONArray = append(settInstJSONArray, stockJSON...)		
	}

	suffix := "]}"
	settInstJSONArray = append(settInstJSONArray, suffix...)

	err = stub.SetEvent("RepoDealClosePosting", []byte(settInstJSONArray))
	if err != nil {
		fmt.Println("SetEvent RepoDealClosePosting Error", err)
	}
	
	return shim.Success([]byte("Repo Deal Close request is captured successfully!"))
}

func (t *RepoDealChaincode) initiateRepoDealClose(stub shim.ChaincodeStubInterface, sysReference string, repoJSON string, tranType string, tranOriginatorParty string, mpChaincodeID string, repoChannelID string) pb.Response {
	fmt.Println("###### RepoDealChaincode: function: initiateRepoDealClose ")

	var err error
	err = mpHandler.newRepoCloseNotification(stub, sysReference, repoJSON, tranType, tranOriginatorParty, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error generating notification for repo deal close ", err)
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Repo Deal Close request is captured successfully!"))
}

func (t *RepoDealChaincode) newCollateralSubUpdate(stub shim.ChaincodeStubInterface, collStruct CollateralDetails, action string) pb.Response {

	fmt.Println("###### RepoDealCC: function: newCollateralSubUpdate ")
	var err error
	if action == "NEW" {
		err = collHandler.newCollateralPosition(stub, collStruct)
		if err != nil {
			return shim.Error(err.Error())
		}

	} else if action == "AMEND" {
		err = collHandler.updateCollateralPosition(stub, collStruct)
		if err != nil {
			return shim.Error(err.Error())
		}

	} else if action == "CANCEL" {
		err = collHandler.deactivateCollateralPosition(stub, collStruct)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success([]byte("Repo captured successfully"))
}

func (t *RepoDealChaincode) initiateDailyInterestCalculation(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("###### RepoDealCC: function: initiateDailyInterestCalculation ")

	var err error
	var arbitrary_json map[string]interface{}

	fmt.Println("Querying for trades for calculating interest")

	tradeListJSON, err := tHandler.queryAllActiveTrade(stub)
	if err != nil {
		fmt.Println("Error querying trade for interest calculation:", err)
		return shim.Error(err.Error())
	}

	fmt.Println("Daily Interest calculation on Trades::", string(tradeListJSON))
	err = json.Unmarshal([]byte(tradeListJSON), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		return shim.Error(err.Error())
	}

	tradeData := arbitrary_json["Trade"].([]interface{})
	fmt.Println("Trade Data: %v", tradeData)

	for key1, value1 := range tradeData {
		fmt.Printf("Trade Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
		jsonByte, err := json.Marshal(value1)
		jsonStr := convertArray(jsonByte)
		fmt.Println("Trade Data JSON Str: %v", jsonStr)
		if value1 != nil {
			var tradeStruct TradeDetails
			err = json.Unmarshal([]byte(jsonByte), &tradeStruct)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}

			if tradeStruct.RepoStatus == "OPENLEGSETTLED" {
				fmt.Println("Calculating interest on the trade :", tradeStruct.ProcessingSystemReference)
				err = tHandler.interestCalculation(stub, tradeStruct)
				if err != nil {
					fmt.Println("Error calculating interest on the trade :", tradeStruct.ProcessingSystemReference, err)
					return shim.Error(err.Error())

				}
			}
		}
	}
	return shim.Success([]byte("Trade level interest calculation successful!"))
}

func (t *RepoDealChaincode) initiateInterimInterestPaymentSettlement(stub shim.ChaincodeStubInterface, sysReference string, interimPayment string, settlementType string, repoChaincodeID string, repoChannelID string) pb.Response {
	var err error
	var settInstJSONArray []byte
	prefix := "{\"Instruction\" : ["

	fmt.Println("###### RepoDealCC: function: initiateInterimInterestPaymentSettlement ")

	postingJSON, err := repHandler.interimInterestPaymentSettlement(stub, sysReference, interimPayment, settlementType, repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error interim payment repo trade :", sysReference, err)
		return shim.Error(err.Error())
	}

	settInstJSONArray = append(settInstJSONArray, prefix...)	
	settInstJSONArray = append(settInstJSONArray, postingJSON...)		
	suffix := "]}"
	settInstJSONArray = append(settInstJSONArray, suffix...)

	err = stub.SetEvent("InterimPaymentPosting", []byte(settInstJSONArray))
	if err != nil {
		fmt.Println("SetEvent InterimPaymentPosting Error", err)
	}

	return shim.Success([]byte("Repo Interim Payment Settlement is successful!"))
}

func (t *RepoDealChaincode) initiateCashAdjustmentSettlement(stub shim.ChaincodeStubInterface, sysReference string, cashPayment string, indicator string, settlementType string, repoChaincodeID string, repoChannelID string) pb.Response {
	var err error
	var settInstJSONArray []byte
	prefix := "{\"Instruction\" : ["

	fmt.Println("###### RepoDealCC: function: initiateCashAdjustmentSettlement ")

	postingJSON, err := repHandler.initiateCashAdjustmentSettlement(stub, sysReference, cashPayment, indicator, settlementType, repoChaincodeID, repoChannelID )
	if err != nil {
		fmt.Println("Error cash adjustment repo trade :", sysReference, err)
		return shim.Error(err.Error())
	}
	settInstJSONArray = append(settInstJSONArray, prefix...)	
	settInstJSONArray = append(settInstJSONArray, postingJSON...)		
	suffix := "]}"
	settInstJSONArray = append(settInstJSONArray, suffix...)

	err = stub.SetEvent("CashAdjustmentPosting", []byte(settInstJSONArray))
	if err != nil {
		fmt.Println("SetEvent CashAdjustmentPosting Error", err)
	}

	return shim.Success([]byte("Repo CashAdjustment Settlement is successful!"))
}

func (t *RepoDealChaincode) initiateRepoDealSettlement(stub shim.ChaincodeStubInterface, sysReference string, repoChaincodeID string, repoChannelID string) pb.Response {

	var err error
	var settlementType string
	var settInstJSONArray []byte
	prefix := "{\"Instruction\" : ["
	var delimiter string
	delimiter = ","
	settlementType = "RepoOpenLegSettlement"
	fmt.Println("###### RepoDealCC: function: initiateRepoDealSettlement ")

	stockJSON, cashJSON , err := repHandler.repoDealSettlement(stub, sysReference, settlementType, repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error settling repo trade :", sysReference, err)
		return shim.Error(err.Error())
	}

	settInstJSONArray = append(settInstJSONArray, prefix...)
	if cashJSON != nil {	
		fmt.Println("cashJSON :", string(cashJSON))	
		settInstJSONArray = append(settInstJSONArray, cashJSON...)
		settInstJSONArray = append(settInstJSONArray, delimiter...)		
	}

	if stockJSON != nil {
		fmt.Println("stockJSON :", string(stockJSON))	
		settInstJSONArray = append(settInstJSONArray, stockJSON...)		
	}

	suffix := "]}"
	settInstJSONArray = append(settInstJSONArray, suffix...)
	
	err = stub.SetEvent("RepoDealOpenPosting", []byte(settInstJSONArray))
	if err != nil {
		fmt.Println("SetEvent RepoDealOpenPosting Error", err)
	}
	fmt.Println("RepoDealOpenPosting :", string(settInstJSONArray))
	
	return shim.Success([]byte("Repo Openleg Settlement initiated successfully"))
}

func (t *RepoDealChaincode) initiateRepoDealCloseSettlement(stub shim.ChaincodeStubInterface, sysReference string, repoChaincodeID string, repoChannelID string) pb.Response {
	var err error
	var settlementType string
	var settInstJSONArray []byte
	prefix := "{\"Instruction\" : ["
	var delimiter string
	delimiter = ","	
	settlementType = "RepoCloseLegSettlement"
	fmt.Println("###### RepoDealCC: function: initiateRepoDealCloseSettlement ")

	stockJSON, cashJSON, err := repHandler.repoDealCloseSettlement(stub, sysReference, settlementType, repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error settling repo trade :", sysReference, err)
		return shim.Error(err.Error())
	}	
	
	settInstJSONArray = append(settInstJSONArray, prefix...)
	if cashJSON != nil {	
		fmt.Println("cashJSON :", string(cashJSON))	
		settInstJSONArray = append(settInstJSONArray, cashJSON...)
		settInstJSONArray = append(settInstJSONArray, delimiter...)		
	}

	if stockJSON != nil {
		fmt.Println("stockJSON :", string(stockJSON))			
		settInstJSONArray = append(settInstJSONArray, stockJSON...)		
	}

	suffix := "]}"
	settInstJSONArray = append(settInstJSONArray, suffix...)
	
	err = stub.SetEvent("RepoDealClosePosting", []byte(settInstJSONArray))
	if err != nil {
		fmt.Println("SetEvent RepoDealClosePosting Error", err)
	}

	return shim.Success([]byte("Repo Close Settlement is successful!"))
}

func (t *RepoDealChaincode) setRepoStatusUpdate(stub shim.ChaincodeStubInterface, sysReference string, newRepoStatus string, sourceSystem string) pb.Response {

	var err error
	fmt.Println("###### RepoDealCC: function: setRepoStatusUpdate ")

	err = repHandler.repoStatusUpdate(stub, sysReference, newRepoStatus, sourceSystem)
	if err != nil {
		fmt.Println("Error updating Repo status ", err)
	}

	return shim.Success([]byte("Repo Status is updated successfully!"))
}

func (t *RepoDealChaincode) queryRepos(stub shim.ChaincodeStubInterface, args string) pb.Response {

	fmt.Println("###### RepoDealCC: function: queryRepos ")

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, args)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func (t *RepoDealChaincode) queryCollaterals(stub shim.ChaincodeStubInterface, borrowerParticipantID string) pb.Response {

	fmt.Println("###### RepoDealCC: function: queryCollaterals ")

	fmt.Println("Querying Repo CC : function queryCollaterals:", borrowerParticipantID)

	if borrowerParticipantID == "" {
		return shim.Error("Incorrect number of arguments. Expecting borrowerParticipantID")
	}

	queryResults, err := collHandler.queryAllCollateralPositionsByParticipant(stub, borrowerParticipantID)
	if err != nil {
		return shim.Error(err.Error())
	}
	if string(queryResults) != "" {
		return shim.Success(queryResults)
	}

	return shim.Success([]byte("EMPTY"))
}

func (t *RepoDealChaincode) queryTradeHistory(stub shim.ChaincodeStubInterface, processingSystemReference string) pb.Response {

	fmt.Println("###### RepoDealCC: function: queryTradeHistory ")

	if processingSystemReference == "" {
		return shim.Error("Incorrect number of arguments. Expecting processingSystemReference")
	}
	//args == ParticipantID
	queryResults, err := tHandler.queryTradeHistoryReport(stub, processingSystemReference)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func (t *RepoDealChaincode) queryAllCollateralPositionsByRepo(stub shim.ChaincodeStubInterface, borrowerParticipantID string, processingSystemReference string) pb.Response {

	fmt.Println("###### RepoDealCC: function: queryAllCollateralPositionsByRepo ")
	fmt.Println("Querying Repo CC : function queryAllCollateralPositionsByRepo:", borrowerParticipantID, processingSystemReference)

	if borrowerParticipantID == "" || processingSystemReference == "" {
		return shim.Error("Incorrect number of arguments. Expecting borrowerParticipantID & processingSystemReference")
	}
	//args == ParticipantID
	queryResults, err := collHandler.queryAllCollateralPositionsByRepo(stub, borrowerParticipantID, processingSystemReference)
	if err != nil {
		return shim.Error(err.Error())
	}
	if string(queryResults) != "" {
		return shim.Success(queryResults)
	}

	return shim.Success([]byte("EMPTY"))
}

func (t *RepoDealChaincode) getRepoDealRepoStatus(stub shim.ChaincodeStubInterface, sysReference string) pb.Response {

	fmt.Println("###### RepoDealCC: function: getRepoDealRepoStatus ")

	if sysReference == "" {
		return shim.Error("Incorrect number of arguments. Expecting sysReference")
	}

	repoStatus, err := repHandler.getRepoStatus(stub, sysReference)
	if err != nil {
		return shim.Error("Error retriving Repo deal info.")
	}

	return shim.Success([]byte(repoStatus))
}

func (t *RepoDealChaincode) getRepoDealInformationWOAC(stub shim.ChaincodeStubInterface, sysReference string) pb.Response {

	fmt.Println("###### RepoDealCC: function: getRepoDealInformationWOAC ")

	if sysReference == "" {
		return shim.Error("Incorrect number of arguments. Expecting sysReference")
	}

	repoDealJSON, err := repHandler.getRepoDealInformation(stub, sysReference)
	if err != nil {
		return shim.Error("Error retriving Repo deal info.")
	}
	return shim.Success(repoDealJSON)
}

func (t *RepoDealChaincode) queryRepoByFilter(stub shim.ChaincodeStubInterface, queryString string) pb.Response {
	//func (t *RepoDealChaincode) queryRepoByFilter(stub shim.ChaincodeStubInterface, queryString string) pb.Response {

	fmt.Println("###### RepoDealCC: function: queryRepoByFilter ")

	//queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"TradeDetails\",\"ActiveInd\":\"A\",\"Counterparty\":\"%s\"}}", counterparty)
	//queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"TradeDetails\",\"s\":\"%s\"}}", counterparty)
	//queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"TradeDetails\",\"ActiveInd\":\"A\"}}")
	//queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"TradeDetails\",\"ActiveInd\":\"A\",\"RepoStatus\":\"OPENLEGSETTLED\", \"MaturityDate\":\"09/20/2017\"}}")
	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *RepoDealChaincode) queryTradeByFilter(stub shim.ChaincodeStubInterface) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryTradeByFilter ")

	queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"TradeDetails\",\"ActiveInd\":\"A\"}}")
	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return nil, err
	}
	return queryResults, nil
}

func (t *RepoDealChaincode) queryLenderCollateralsByFilter(stub shim.ChaincodeStubInterface, participant string) pb.Response {

	fmt.Println("###### RepoDealCC: function: queryLenderCollateralsByFilter ")

	queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"CollateralDetails\",\"ActiveInd\":\"A\",\"LenderParticipantID\":\"%s\"}}", participant)
	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *RepoDealChaincode) queryBorrowerCollateralsByFilter(stub shim.ChaincodeStubInterface, participant string) pb.Response {

	fmt.Println("###### RepoDealCC: function: queryBorrowerCollateralsByFilter ")

	queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"CollateralDetails\",\"ActiveInd\":\"A\",\"BorrowerParticipantID\":\"%s\"}}", participant)
	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *RepoDealChaincode) queryCollateralPositionByInstrument(stub shim.ChaincodeStubInterface, participantID string, sysReference string, instrumentID string) pb.Response {
	fmt.Println("###### RepoDealCC: function: queryCollateralPositionByInstrument ")

	queryResults, err := collHandler.queryCollateralPositionByInstrument(stub, participantID, sysReference, instrumentID)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *RepoDealChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("###### RepoDealCC: function: Invoke ")

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

func (t *RepoDealChaincode) invoke(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {

	fmt.Println("###### RepoDealCC: function: invoke ")

	if function == "" {
		return shim.Error("Invoke function is not passed")
	}

	var err error
	var jsonData string
	var action string
	var multiChaincodeID string
	var repoChannelID string
	var arbitrary_json map[string]interface{}

	if function == "repoDealCapture" {
		jsonData = args[1]
		action = args[2]
		multiChaincodeID = args[3]
		repoChannelID = args[4]
		return t.repoDealCapture(stub, jsonData, action, multiChaincodeID, repoChannelID)

	} else if function == "newCollateralSubUpdate" {
		jsonData = args[1]
		action = args[2]

		var collStruct CollateralDetails
		err = json.Unmarshal([]byte(jsonData), &collStruct)
		if err != nil {
			fmt.Println("Error parsing JSON: ", err)
		}

		return t.newCollateralSubUpdate(stub, collStruct, action)

	} else if function == "newCollateralValuationUpdate" {
		jsonData = args[1]
		action = args[2]
		fmt.Println("Collateral Valuation Data JSON received: %v", jsonData)
		err = json.Unmarshal([]byte(jsonData), &arbitrary_json)
		if err != nil {
			fmt.Println("Error parsing JSON: ", err)
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
				}

				err = collHandler.newCollateralValUpdate(stub, collStruct, action)
				if err != nil {
					fmt.Println("Error updating collateral for valuation: ", err)
				}
			}
		}
		return shim.Success([]byte("Collateral valuation Updated successfully!"))

	} else if function == "repoDealApproval" {
		sysReference := args[1]
		repoJSON := args[2]
		notificationType := args[3]
		multiPartyAction := args[4]
		return t.repoDealApproval(stub, sysReference, repoJSON, notificationType, multiPartyAction)

	} else if function == "initiateRepoDealSettlement" {
		sysReference := args[1]		
		repoChaincodeID := args[2]
		repoChannelID := args[3]

		return t.initiateRepoDealSettlement(stub, sysReference, repoChaincodeID, repoChannelID)

	} else if function == "initiateInterestPaymentRequest" {
		sysReference := args[1]
		payment := args[2]
		party := args[3]
		counterparty := args[4]
		tranType := "InterestPayment"
		tranOriginatorParty := args[6]
		mpChaincodeID := args[7]
		repoChannelID := args[8]
		return t.initiateInterestPaymentRequest(stub, sysReference, payment, party, counterparty, tranType, tranOriginatorParty, mpChaincodeID, repoChannelID)

	} else if function == "initiateInterimInterestPaymentSettlement" {
		sysReference := args[1]
		payment := args[2]
		settlementType := args[3]
		repoChaincodeID := args[4]
		repoChannelID := args[5]
		return t.initiateInterimInterestPaymentSettlement(stub, sysReference, payment, settlementType, repoChaincodeID, repoChannelID)

	} else if function == "initiateRepoDealClose" {
		sysReference := args[1]
		repoJSON := args[2]
		tranType := args[3]
		tranOriginatorParty := args[4]
		mpChaincodeID := args[5]
		repoChannelID := args[6]
		return t.initiateRepoDealClose(stub, sysReference, repoJSON, tranType, tranOriginatorParty, mpChaincodeID, repoChannelID)

	} else if function == "initiateRepoDealCloseSettlement" {
		sysReference := args[1]		
		repoChaincodeID := args[2]
		repoChannelID := args[3]
		return t.initiateRepoDealCloseSettlement(stub, sysReference, repoChaincodeID, repoChannelID)

	} else if function == "initiateAutoRepoDealClose" {
		sysReference := args[1]
		repoChaincodeID := args[2]
		repoChannelID := args[3]
		return t.initiateAutoRepoDealClose(stub, sysReference, repoChaincodeID, repoChannelID)

	} else if function == "initiateDailyInterestCalculation" {

		return t.initiateDailyInterestCalculation(stub)
	} else if function == "setRepoStatusUpdate" {
		sysReference := args[1]
		newRepoStatus := args[2]
		sourceSystem := args[3]

		return t.setRepoStatusUpdate(stub, sysReference, newRepoStatus, sourceSystem)
	} else if function == "repoCollateralSubCapture" {
		sysReference := args[1]
		jsonData := args[2]
		action := args[3]
		mpChaincodeID := args[4]
		repoChannelID := args[5]

		return t.repoCollateralSubCapture(stub, sysReference, jsonData, action, mpChaincodeID, repoChannelID)
	} else if function == "initiateRepoDealCollateralSub" {
		sysReference := args[1]
		jsonData := args[2]
		repoChaincodeID := args[3]
		repoChannelID := args[4]

		return t.initiateRepoDealCollateralSub(stub, sysReference, jsonData, repoChaincodeID, repoChannelID)
	} else if function == "initiateCashAdjustmentRequest" {
		sysReference := args[1]
		payment := args[2]
		indicator := args[3]
		party := args[4]
		counterparty := args[5]
		tranType := "CashAdjustment"
		tranOriginatorParty := args[7]
		mpChaincodeID := args[8]
		repoChannelID := args[9]

		return t.initiateCashAdjustmentRequest(stub, sysReference, payment, indicator, party, counterparty, tranType, tranOriginatorParty, mpChaincodeID, repoChannelID)

	} else if function == "initiateCashAdjustmentSettlement" {
		sysReference := args[1]
		payment := args[2]
		indicator := args[3]
		settlementType := args[4]
		repoChaincodeID := args[5]
		repoChannelID := args[6]
		return t.initiateCashAdjustmentSettlement(stub, sysReference, payment, indicator, settlementType, repoChaincodeID, repoChannelID)

	}
	return shim.Error("Received unknown function invocation")
}

func (t *RepoDealChaincode) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### RepoDealCC: function: query ")
	fmt.Println("[RepoDealChaincode] Query")

	if function == "" || len(args) < 2 {
		return shim.Error("Not enough args passed")
	}

	if function == "queryRepos" {
		var ref = args[1]
		fmt.Println("Querying for %v", ref)
		return t.queryRepos(stub, ref)

	} else if function == "getRepoDealInformationWOAC" {
		var ref = args[1]
		fmt.Println("Querying for %v", ref)
		return t.getRepoDealInformationWOAC(stub, ref)

	} else if function == "queryCollaterals" {
		var ref = args[1]
		fmt.Println("Querying for %v", ref)
		return t.queryCollaterals(stub, ref)

	} else if function == "queryTradeHistory" {
		var ref = args[1]
		fmt.Println("Querying for %v", ref)
		return t.queryTradeHistory(stub, ref)

	} else if function == "queryAllCollateralPositionsByRepo" {
		var borrowerPaticipatnID = args[1]
		var ref = args[2]
		fmt.Println("Querying for %v", borrowerPaticipatnID, ref)
		return t.queryAllCollateralPositionsByRepo(stub, borrowerPaticipatnID, ref)

	} else if function == "queryRepoByFilter" {
		var queryString = args[1]
		return t.queryRepoByFilter(stub, queryString)

	} else if function == "getRepoDealRepoStatus" {
		var tradeRef = args[1]
		return t.getRepoDealRepoStatus(stub, tradeRef)

	} else if function == "queryLenderCollateralsByFilter" {
		var participantID = args[1]
		return t.queryLenderCollateralsByFilter(stub, participantID)

	} else if function == "queryBorrowerCollateralsByFilter" {
		var participantID = args[1]
		return t.queryBorrowerCollateralsByFilter(stub, participantID)

	} else if function == "queryCollateralPositionByInstrument" {
		var participantID = args[1]
		var sysReference = args[2]
		var instrumentID = args[3]
		return t.queryCollateralPositionByInstrument(stub, participantID, sysReference, instrumentID)
	}

	return shim.Error("Received unknown function query invocation with function")
}

func (t *RepoDealChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("###### RepoDealCC: function: Init ")

	_, args := stub.GetFunctionAndParameters()
	fmt.Println("[RepoDealChaincode] Init")
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
	fmt.Println("###### RepoDealCC: function: main ")
	//	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(RepoDealChaincode))
	

	if err != nil {
		fmt.Println("Error starting RepoDealChaincode: %s", err)
	}
}
