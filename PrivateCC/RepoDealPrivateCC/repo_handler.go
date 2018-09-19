package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//repoHandler provides APIs used to perform operations on CC's KV store
type repoHandler struct {
}

// NewPostingHandler create a new postings
func NewRepoHandler() *repoHandler {
	return &repoHandler{}
}

func (t *repoHandler) enrichCollateral(stub shim.ChaincodeStubInterface, collStruct CollateralDetails) (float64, float64, error) {
	var quantity float64
	var cleanPrice float64
	var principal float64
	var dirtyPrice float64
	var haircut float64
	var net float64

	fmt.Println("###### RepoDealCC: function: enrichCollateral ")

	if collStruct.Principal == "" && collStruct.Quantity != "" && collStruct.CleanPrice != "" {
		quantity, _ = strconv.ParseFloat(collStruct.Quantity, 64)
		cleanPrice, _ = strconv.ParseFloat(collStruct.CleanPrice, 64)
		principal = quantity * cleanPrice
		//collStruct.Principal = strconv.FormatFloat(principal, 'f', 2, 64)
	} else {
		return 0.0, 0.0, errors.New("Error in enriching Principal field: not enough parameters passed!")
	}

	if collStruct.NetConsiderationBaseCurrency == "" && collStruct.Quantity != "" && collStruct.DirtyPrice != "" && collStruct.Haircut != "" {
		quantity, _ = strconv.ParseFloat(collStruct.Quantity, 64)
		dirtyPrice, _ = strconv.ParseFloat(collStruct.DirtyPrice, 64)
		haircut, _ = strconv.ParseFloat(collStruct.Haircut, 64)
		net = quantity * dirtyPrice * 1 / haircut * 1 / 100
		//collStruct.NetConsiderationBaseCurrency = strconv.FormatFloat(net, 'f', 2, 64)
	} else {
		return 0.0, 0.0, errors.New("Error in enriching Net consideration field: not enough parameters passed!")
	}

	return principal, net, nil
}

func (t *repoHandler) repoStatusUpdate(stub shim.ChaincodeStubInterface, sysReference string, newRepoStatus string, sourceSystem string) error {

	var err error
	fmt.Println("###### RepoDealCC: function: repoStatusUpdate ")

	if sysReference == "" || newRepoStatus == "" {
		return errors.New("Not enough arguments passed")
	}
	fmt.Println("Querying trade for Repo status Update  :", sysReference)

	tradeJSON, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Error querying active trade:", sysReference, err)
		return err
	}

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeJSON), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}

	tradeStruct.RepoStatus = newRepoStatus
	tradeStruct.LastUpdatedUser = sourceSystem

	err = tHandler.updateTrade(stub, tradeStruct)
	if err != nil {
		fmt.Println("Error updating repo status on the trade :", sysReference, err)
		return err
	}
	return nil
}

func (t *repoHandler) getRepoStatus(stub shim.ChaincodeStubInterface, sysReference string) (string, error) {

	fmt.Println("###### RepoDealCC: function: getRepoStatus ")

	if sysReference == "" {
		return "nil", errors.New("Not enough args passed")
	}

	tradeJSON, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		return "nil", err
	}
	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeJSON), &tradeStruct)
	if err != nil {
		fmt.Println("Error unmarshalling trade JSON", err)
	}

	return tradeStruct.RepoStatus, nil
}


func (t *repoHandler) getRepoDealInformation(stub shim.ChaincodeStubInterface, sysReference string) ([]byte, error) {
	
	fmt.Println("###### RepoDealCC: function: getRepoDealInformation")
	
	var repoStruct RepoDeal
	var tradeStruct TradeDetails
	var partyStruct ParticipantDetails
	var partyCustodianStruct ParticipantDetails
	var cpartyStruct ParticipantDetails
	var cpartyCustodianStruct ParticipantDetails
	var collArray []CollateralDetails
	var partyRow []byte
	var partyCustodianRow []byte
	var counterpartyRow []byte
	var counterpartyCustodianRow []byte
	var collRow []byte
	var repoDealJSON []byte
	var err error
	var colllender string 
	var collborrower string
	var collBorrowerRow []byte
	var collLenderRow []byte
	
	fmt.Println("Querying for : [%s]", string(sysReference))
	
	if sysReference != "" {
		//trade
		tradeRow, err := tHandler.queryActiveTrade(stub, sysReference)
		if err != nil {
			fmt.Println("Failed to query a position row [%s]", err)
		}
		fmt.Println("Trade Information:", string(tradeRow))
	
		if string(tradeRow) != "" {
			err = json.Unmarshal([]byte(tradeRow), &tradeStruct)
			if err != nil {
				fmt.Println("Error parsing tradeDetails JSON: ", err)
			}
			repoStruct.Trade = tradeStruct
		}
	}
	
	//party
	if tradeStruct.Party != "" {
		fmt.Println("Querying for party :", string(tradeStruct.Party))
	
		partyRow, err = partHandler.queryParticipant(stub, tradeStruct.Party, sysReference)
		if err != nil {
			fmt.Println("Failed to query a party row [%s]", err)
		}
	
		if string(partyRow) != "" {
			err = json.Unmarshal([]byte(partyRow), &partyStruct)
			if err != nil {
				fmt.Println("Error parsing partyDetails JSON: ", err)
			}
			repoStruct.Party = partyStruct
			if partyStruct.PartyType == "COLLLENDER" {
				colllender = partyStruct.ParticipantID 
			} else {
				collborrower = partyStruct.ParticipantID
			}
		}
	}
	
	//party custodian
	if tradeStruct.PartyCustodian != "" {
		fmt.Println("Querying for PartyCustodian :", string(tradeStruct.PartyCustodian))
		partyCustodianRow, err = partHandler.queryParticipant(stub, tradeStruct.PartyCustodian, sysReference)
		if err != nil {
			fmt.Println("Failed to query a partyCustodian row [%s]", err)
		}
	
		if string(partyCustodianRow) != "" {
			err = json.Unmarshal([]byte(partyCustodianRow), &partyCustodianStruct)
			if err != nil {
				fmt.Println("Error parsing party Custodian Details JSON: ", err)
			}
			repoStruct.PartyCustodian = partyCustodianStruct
		}
	}
	
	//Counterparty
	if tradeStruct.Counterparty != "" {
		fmt.Println("Querying for counterparty:", string(tradeStruct.Counterparty))
		counterpartyRow, err = partHandler.queryParticipant(stub, tradeStruct.Counterparty, sysReference)
		if err != nil {
			fmt.Println("Failed to query a counterparty row [%s]", err)
		}
	
		if string(counterpartyRow) != "" {
			err = json.Unmarshal([]byte(counterpartyRow), &cpartyStruct)
			if err != nil {
				fmt.Println("Error parsing cpartyDetails JSON: ", err)
			}
			repoStruct.Counterparty = cpartyStruct
			if cpartyStruct.PartyType == "COLLLENDER" {
				colllender = cpartyStruct.ParticipantID
			} else {
				collborrower = cpartyStruct.ParticipantID
			}
		}
	}
	
	//counterparty custodian
	if tradeStruct.CounterpartyCustodian != "" {
		fmt.Println("Querying for CounterpartyCustodian :", string(tradeStruct.CounterpartyCustodian))
		counterpartyCustodianRow, err = partHandler.queryParticipant(stub, tradeStruct.CounterpartyCustodian, sysReference)
		if err != nil {
			fmt.Println("Failed to query a counterpartyCustodian row [%s]", err)
		}
	
		if string(counterpartyCustodianRow) != "" {
			err = json.Unmarshal([]byte(counterpartyCustodianRow), &cpartyCustodianStruct)
			if err != nil {
				fmt.Println("Error parsing cparty custodian Details JSON: ", err)
			}
			repoStruct.CounterpartyCustodian = cpartyCustodianStruct
		}
	}
	
	if tradeStruct.TradeType == "REPURCHASE" || tradeStruct.TradeType == "REVERSEREPURCHASE" || tradeStruct.TradeType == "PLEDGEBORROW" {
		collBorrowerRow, err = collHandler.queryAllCollateralPositionsByRepo(stub, collborrower, sysReference)
		if err != nil {
			fmt.Println("Failed to query a collateral position row [%s]", err)
		}
	}
	
	if tradeStruct.TradeType == "PLEDGEBORROW" {
		collLenderRow, err = collHandler.queryAllCollateralPositionsByRepo(stub, colllender, sysReference)
		if err != nil {
			fmt.Println("Failed to query a collateral position row [%s]", err)
		}
	}
	
	fmt.Println("Collateral Information %s", string(collRow))
	
	if string(collBorrowerRow) != "" || string(collLenderRow) != "" {
		// Read multiple collaterals
		if string(collBorrowerRow) != "" {
			var arbitrary_json map[string]interface{}
			err = json.Unmarshal(collBorrowerRow, &arbitrary_json)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}
	
			collateralData := arbitrary_json["Collateral"].([]interface{})
			fmt.Println("Collateral Data: %v", collateralData)
	
			for key1, value1 := range collateralData {
				fmt.Printf("Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
				jsonByte, err := json.Marshal(value1)
	
				var collStruct CollateralDetails
				err = json.Unmarshal([]byte(jsonByte), &collStruct)
				if err != nil {
					fmt.Println("Error parsing collDetails JSON: ", err)
				}
	
				collArray = append(collArray, collStruct)
			}
		}
			
		if string(collLenderRow) != "" {
			var arbitrary_json map[string]interface{}
			err = json.Unmarshal(collLenderRow, &arbitrary_json)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}
	
			collateralData := arbitrary_json["Collateral"].([]interface{})
			fmt.Println("Collateral Data: %v", collateralData)
	
			for key1, value1 := range collateralData {
	
				fmt.Printf("Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
				jsonByte, err := json.Marshal(value1)
	
				var collStruct CollateralDetails
				err = json.Unmarshal([]byte(jsonByte), &collStruct)
				if err != nil {
					fmt.Println("Error parsing collDetails JSON: ", err)
				}
				collArray = append(collArray, collStruct)
			}
		}
	
		repoStruct.Collaterals = collArray
	}
	
	repoDealJSON, err = json.Marshal(repoStruct)
	return []byte(repoDealJSON), err
}


func (t *repoHandler) repoDealDeployment(stub shim.ChaincodeStubInterface, systemReference string, repoJSON string, action string, repoStatus string) (error) {
	
	fmt.Println("###### RepoDealCC: function: repoDealDeployment ")
	
	if systemReference == "" || repoJSON == "" || action == "" {
		shim.Error("Enough args are not passed")
	}
	
	var err error
	var principal float64
	var netConsideration float64
	var arbitrary_json map[string]interface{}
	
	fmt.Println("Repo Data JSON received: %v", repoJSON)
	err = json.Unmarshal([]byte(repoJSON), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}

	//Capture trade details
	var tradeStruct TradeDetails
	jsonByte, err := json.Marshal(arbitrary_json["Trade"])
	jsonStr := convertArray(jsonByte)
	fmt.Println("Trade Data JSON Str: %v", jsonStr)
	if string(jsonStr) != "null" {
		err = json.Unmarshal([]byte(jsonStr), &tradeStruct)
		if err != nil {
			fmt.Println("Error parsing JSON: ", err)
		}
	
		if action == "NEW" {
	
			tradeStruct.ProcessingSystemReference = systemReference //ONLY FOR NEW	
			tradeStruct.RepoStatus = repoStatus
	
		} else if action == "AMEND" || action == "CANCEL" {
	
			fmt.Println("Amend or Cancel for transaction", action)
			if tradeStruct.ProcessingSystemReference != "" {
				exJSONbyte, err := tHandler.queryActiveTrade(stub, tradeStruct.ProcessingSystemReference)
				if err != nil || string(exJSONbyte) == "" {
					fmt.Println("Active trade for AMEND OR CANCEL is not found!")
					return err
				}
				
				var exTradeStruct TradeDetails
				if string(exJSONbyte) != "null" {
					err = json.Unmarshal([]byte(exJSONbyte), &exTradeStruct)
					if err != nil {
						fmt.Println("Error parsing JSON: ", err)
					}				
				}
				tradeStruct.RepoStatus = repoStatus
				tradeStruct.Version = exTradeStruct.Version
	
			} else {
					fmt.Println("ProcessingSystemReference is not supplited for AMEND or CANCEL")
					return errors.New("ProcessingSystemReference is not supplited for AMEND or CANCEL")
			}
		}
		
		err = tHandler.newTradeCapture(stub, tradeStruct, action)
		if err != nil {
			fmt.Println("Trade Capture failed for: ", arbitrary_json["Trade"].(string), err)
			return err
		}
	}
	
	//Capture party 1
	jsonByte, err = json.Marshal(arbitrary_json["Party"])
	jsonStr = convertArray(jsonByte)
	fmt.Println("Party Data JSON Str: %v", jsonStr)
	
	if string(jsonStr) != "null" {
		var partyStruct ParticipantDetails
		err = json.Unmarshal([]byte(jsonStr), &partyStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
	
		if action == "NEW" {
			partyStruct.ProcessingSystemReference = systemReference //ONLY FOR NEW
		}
	
		err = partHandler.newParticipantCapture(stub, partyStruct, "Party", action)
		if err != nil {
			fmt.Println("Party Capture failed for: %v : %v", arbitrary_json["Party"].(string), err)		
			return err
		}
			/*
			if partyStruct.PartyType == "COLLLENDER" {
				colllender = partyStruct.ParticipantID
			} else {
				collborrower = partyStruct.ParticipantID
			}
			*/
	}
	
	//Capture custodian for party 1
	jsonByte, err = json.Marshal(arbitrary_json["CustodianParty"])
	jsonStr = convertArray(jsonByte)
	fmt.Println("Custodian Data JSON Str: %v", jsonStr)
	if string(jsonStr) != "null" {
		var partyStruct ParticipantDetails
		err = json.Unmarshal([]byte(jsonStr), &partyStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
	
		if action == "NEW" {
			partyStruct.ProcessingSystemReference = systemReference //ONLY FOR NEW
		}
	
		err = partHandler.newParticipantCapture(stub, partyStruct, "Custodian", action)
		if err != nil {
			fmt.Println("Custodian Capture failed for: %v : %v", arbitrary_json["Custodian"].(string), err)	
			return err
		}
	}
	
	//Capture party 2
	jsonByte, err = json.Marshal(arbitrary_json["Counterparty"])
	jsonStr = convertArray(jsonByte)
	fmt.Println("Counterparty Data JSON Str: %v", jsonStr)
	if string(jsonStr) != "null" {
		var partyStruct ParticipantDetails
		err = json.Unmarshal([]byte(jsonStr), &partyStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
	
		if action == "NEW" {
			partyStruct.ProcessingSystemReference = systemReference //ONLY FOR NEW
		}
	
		err = partHandler.newParticipantCapture(stub, partyStruct, "Counterparty", action)
		if err != nil {
			fmt.Println("Counterparty Capture failed for: %v : %v", arbitrary_json["Counterparty"].(string), err)	
			return err
		}
		/*
		if partyStruct.PartyType == "COLLLENDER" {
			colllender = partyStruct.ParticipantID
		} else {
			collborrower = partyStruct.ParticipantID
		}
		*/
	}
	
	//Capture custodian for party 2
	jsonByte, err = json.Marshal(arbitrary_json["CustodianCounterparty"])
	jsonStr = convertArray(jsonByte)
	fmt.Println("Custodian Data JSON Str: %v", jsonStr)
	if string(jsonStr) != "null" {
		var partyStruct ParticipantDetails
		err = json.Unmarshal([]byte(jsonStr), &partyStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
	
		if action == "NEW" {
			partyStruct.ProcessingSystemReference = systemReference //ONLY FOR NEW
		}
	
		err = partHandler.newParticipantCapture(stub, partyStruct, "Custodian", action)
		if err != nil {
			fmt.Println("Custodian Capture failed for: %v : %v", arbitrary_json["Custodian"].(string), err)		
			return err
		}
	}
	
	//Capture Multiple Collaterals
	collateralData := arbitrary_json["Collateral"].([]interface{})
	fmt.Println("Collateral Data: %v", collateralData)

	for key1, value1 := range collateralData {
		fmt.Printf("Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
		jsonByte, err = json.Marshal(value1)
		jsonStr = convertArray(jsonByte)
		fmt.Println("Collateral Data JSON Str: %v", jsonStr)

		if value1 != nil {
			var collStruct CollateralDetails
			err = json.Unmarshal([]byte(jsonStr), &collStruct)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}
	
			if action == "NEW" {
				collStruct.ProcessingSystemReference = systemReference //ONLY FOR NEW
				// UI is setting these fields
				//collStruct.LenderParticipantID = colllender
				//collStruct.BorrowerParticipantID = collborrower
				collStruct.TransactionDate = tradeStruct.TransactionDate
				collStruct.TransactionTimestamp = tradeStruct.TransactionTimestamp
				collStruct.EffectiveDate = tradeStruct.EffectiveDate
				collStruct.ContractualValueDate = tradeStruct.ContractualValueDate
				collStruct.CloseEventDate = tradeStruct.CloseEventDate
			}

			principal, netConsideration, err = repHandler.enrichCollateral(stub, collStruct)
			if err != nil {
				fmt.Println("Error enriching collateral: ", err)
			}
			fmt.Println("Enriched Principal, NetConsideration::", principal, netConsideration)
			if principal != 0.00 {
				collStruct.Principal = strconv.FormatFloat(principal, 'f', 2, 64)
			}
			if netConsideration != 0.00 {
				collStruct.NetConsiderationBaseCurrency = strconv.FormatFloat(netConsideration, 'f', 2, 64)
			}
			err = collHandler.newCollateralCapture(stub, collStruct, action)
			if err != nil {
				fmt.Println("Collateral Capture failed for: %v : %v", arbitrary_json["Collateral"].(string), err)
				//return shim.Error(err.Error())
				return err
			}
	
		}
	}
	return nil
}

func (t *repoHandler) repoDealSettlement(stub shim.ChaincodeStubInterface, sysReference string, settlementType string, repoChaincodeID string, repoChannelID string) ([]byte, []byte, error) {
	
	var err error
	var colllender ParticipantDetails
	var collborrower ParticipantDetails
	var assetTranArray []AssetTransfers
	var cashTranArray []AssetTransfers
	var settInstJSON []byte	
	
	fmt.Println("###### RepoDealCC: function: repoDealSettlement ")

	tradeJSON, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo trade information", err)
		return nil, nil, err
	}

	fmt.Println("repoDealSettlement: Settling trade JSON:", string(tradeJSON))

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeJSON), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing trade JSON: ", err)
	}

	partyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Party, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo party information", err)
	}

	cpartyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Counterparty, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo counterparty information", err)
	}

	var partyStruct ParticipantDetails
	err = json.Unmarshal([]byte(partyJSON), &partyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling party JSON: ", err)
	}

	if partyStruct.PartyType == "COLLLENDER" {
		colllender = partyStruct
	} else {
		collborrower = partyStruct
	}

	var cpartyStruct ParticipantDetails
	err = json.Unmarshal([]byte(cpartyJSON), &cpartyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling cpty JSON: ", err)
	}

	if cpartyStruct.PartyType == "COLLLENDER" {
		colllender = cpartyStruct
	} else {
		collborrower = cpartyStruct
	}

	if tradeStruct.TradeType == "REPURCHASE" || tradeStruct.TradeType == "REVERSEREPURCHASE" || tradeStruct.TradeType == "PLEDGEBORROW" || tradeStruct.TradeType == "PLEDGEBORROW" {
		
		collBorrowerRow, err := collHandler.queryAllCollateralPositionsByRepo(stub, collborrower.ParticipantID, tradeStruct.ProcessingSystemReference)
		if err != nil {
			fmt.Println("Error retriving Repo Collaterals information", err)
			return nil, nil, err
		}
	
		fmt.Println("repoDealSettlement: Collaterals Borrower JSON", string(collBorrowerRow))
		if string(collBorrowerRow) != "" {
			// Read multiple collaterals
			var carbitrary_json map[string]interface{}
			err = json.Unmarshal([]byte(collBorrowerRow), &carbitrary_json)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}

			collateralData := carbitrary_json["Collateral"].([]interface{})
			fmt.Println("repoDealSettlement: Collateral Data: %v", collateralData)

			for key1, value1 := range collateralData {
				var assetTranStruct AssetTransfers
				
				fmt.Printf("Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
				jsonByte, err := json.Marshal(value1)
				//jsonStr := convertArray(jsonByte)
				fmt.Println("repoDealSettlement: Collateral Data JSON Str: %v", string(jsonByte))

				var collStruct CollateralDetails
				err = json.Unmarshal([]byte(jsonByte), &collStruct)
				if err != nil {
					fmt.Println("Error retriving Repo Collaterals information", err)
					fmt.Println("Error parsing collDetails JSON: ", err)
				}
				
				if collStruct.SubAccount != "" {
					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.LenderParticipantID
					assetTranStruct.FromAcct = collStruct.LenderParticipantAcct
					assetTranStruct.ToParty = collStruct.BorrowerParticipantID
					assetTranStruct.ToAcct = collStruct.SubAccount			
			
				} else {	
					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.LenderParticipantID
					assetTranStruct.FromAcct = collStruct.LenderParticipantAcct
					assetTranStruct.ToParty = collStruct.BorrowerParticipantID
					assetTranStruct.ToAcct = collStruct.BorrowerParticipantAcct			
			
				}
				assetTranArray = append(assetTranArray, assetTranStruct)
			}
		}
		
		settInstJSON, err = instHandler.newSettInstructionEntry(stub, sysReference, tradeStruct.TradeType, assetTranArray, "OpenLeg Settlement", settlementType, "STOCK", repoChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error generating settlement instruction", err)
			return nil, nil, err
		}

	}

	if tradeStruct.TradeType == "PLEDGEBORROW" {

		collLenderRow, err := collHandler.queryAllCollateralPositionsByRepo(stub, colllender.ParticipantID, tradeStruct.ProcessingSystemReference)
		if err != nil {
			fmt.Println("Failed to query a collateral position row [%s]", err)
		}

		fmt.Println("repoDealSettlement: Collaterals Lender JSON", string(collLenderRow))
		if string(collLenderRow) != "" {
			// Read multiple collaterals
			var carbitrary_json map[string]interface{}
			err = json.Unmarshal([]byte(collLenderRow), &carbitrary_json)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}

			collateralData := carbitrary_json["Collateral"].([]interface{})
			fmt.Println("repoDealSettlement: Collateral Data: %v", collateralData)

			for key1, value1 := range collateralData {
				var assetTranStruct AssetTransfers
				
				fmt.Printf("Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
				jsonByte, err := json.Marshal(value1)
				//jsonStr := convertArray(jsonByte)
				fmt.Println("repoDealSettlement: Collateral Data JSON Str: %v", string(jsonByte))

				var collStruct CollateralDetails
				err = json.Unmarshal([]byte(jsonByte), &collStruct)
				if err != nil {
					fmt.Println("Error retriving Repo Collaterals information", err)
					fmt.Println("Error parsing collDetails JSON: ", err)
				}
				
				if collStruct.SubAccount != "" {
					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.LenderParticipantID
					assetTranStruct.FromAcct = collStruct.LenderParticipantAcct
					assetTranStruct.ToParty = collStruct.BorrowerParticipantID
					assetTranStruct.ToAcct = collStruct.SubAccount			
			
				} else {	
					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.LenderParticipantID
					assetTranStruct.FromAcct = collStruct.LenderParticipantAcct
					assetTranStruct.ToParty = collStruct.BorrowerParticipantID
					assetTranStruct.ToAcct = collStruct.BorrowerParticipantAcct			
			
				}

				assetTranArray = append(assetTranArray, assetTranStruct)
			}
		}
	} 

	if tradeStruct.TradeType == "REPURCHASE" || tradeStruct.TradeType == "REVERSEREPURCHASE" {
	
		var cashTranStruct AssetTransfers
		cashTranStruct.InstrumentID = tradeStruct.SettleCurrency
		cashTranStruct.PositionQty = tradeStruct.TotalCashAmount
		cashTranStruct.FromParty = collborrower.ParticipantID
		cashTranStruct.FromAcct = collborrower.TradingAccount
		cashTranStruct.ToParty = colllender.ParticipantID
		cashTranStruct.ToAcct = colllender.TradingAccount

		cashTranArray = append(cashTranArray, cashTranStruct)

		cashInstJSON, err := instHandler.newSettInstructionEntry(stub, sysReference, tradeStruct.TradeType, cashTranArray, "OpenLeg Settlement", settlementType, "CASH", repoChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error generating settlement instruction", err)
			return nil, nil, err
		}
		return []byte(settInstJSON), []byte(cashInstJSON), nil
	}	

	return []byte(settInstJSON), nil, nil
}

func (t *repoHandler) repoDealCloseSettlement(stub shim.ChaincodeStubInterface, sysReference string, settlementType string, repoChaincodeID string, repoChannelID string) ([]byte, []byte, error) {

	var err error
	var colllender ParticipantDetails
	var collborrower ParticipantDetails
	var assetTranArray []AssetTransfers
	var cashTranArray []AssetTransfers
	var settInstJSON []byte
	
	fmt.Println("###### RepoDealCC: function: repoDealCloseSettlement ")

	tradeJSON, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo trade information", err)
		return nil, nil, err
	}

	fmt.Println("repoDealCloseSettlement: Settling trade JSON:", string(tradeJSON))

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeJSON), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing trade JSON: ", err)
	}

	partyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Party, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo party information", err)
	}

	cpartyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Counterparty, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo counterparty information", err)
	}

	var partyStruct ParticipantDetails
	err = json.Unmarshal([]byte(partyJSON), &partyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling party JSON: ", err)
	}

	if partyStruct.PartyType == "COLLLENDER" {
		colllender = partyStruct
	} else {
		collborrower = partyStruct
	}

	var cpartyStruct ParticipantDetails
	err = json.Unmarshal([]byte(cpartyJSON), &cpartyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling cpty JSON: ", err)
	}

	if cpartyStruct.PartyType == "COLLLENDER" {
		colllender = cpartyStruct
	} else {
		collborrower = cpartyStruct
	}

	if tradeStruct.TradeType == "REPURCHASE" || tradeStruct.TradeType == "REVERSEREPURCHASE" || tradeStruct.TradeType == "PLEDGEBORROW"   {
		
		collBorrowerRow, err := collHandler.queryAllCollateralPositionsByRepo(stub, collborrower.ParticipantID, tradeStruct.ProcessingSystemReference)
		if err != nil {
			fmt.Println("Error retriving Repo Collaterals information", err)
			return nil, nil, err
		}

		if string(collBorrowerRow) != "" {
			
			// Read multiple collaterals
			var carbitrary_json map[string]interface{}
			err = json.Unmarshal(collBorrowerRow, &carbitrary_json)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}

			collateralData := carbitrary_json["Collateral"].([]interface{})
			fmt.Println("repoDealCloseSettlement: Collateral Data: %v", collateralData)

			for key1, value1 := range collateralData {
				var assetTranStruct AssetTransfers

				fmt.Printf("repoDealCloseSettlement: Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
				jsonByte, err := json.Marshal(value1)
				jsonStr := convertArray(jsonByte)
				fmt.Println("Collateral Data JSON Str: %v", jsonStr)

				var collStruct CollateralDetails
				err = json.Unmarshal([]byte(jsonByte), &collStruct)
				if err != nil {
					fmt.Println("Error retriving Repo Collaterals information", err)
					fmt.Println("Error parsing collDetails JSON: ", err)
				}

				if collStruct.SubAccount != "" {

					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.BorrowerParticipantID
					assetTranStruct.FromAcct = collStruct.SubAccount
					assetTranStruct.ToParty = collStruct.LenderParticipantID
					assetTranStruct.ToAcct = collStruct.LenderParticipantAcct	
			
				} else {

					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.BorrowerParticipantID
					assetTranStruct.FromAcct = collStruct.BorrowerParticipantAcct
					assetTranStruct.ToParty = collStruct.LenderParticipantID
					assetTranStruct.ToAcct = collStruct.LenderParticipantAcct

				}
				assetTranArray = append(assetTranArray, assetTranStruct)		
			}
		}
	}

	if tradeStruct.TradeType == "PLEDGEBORROW" {
		
		collLenderRow, err := collHandler.queryAllCollateralPositionsByRepo(stub, colllender.ParticipantID, tradeStruct.ProcessingSystemReference)
		if err != nil {
			fmt.Println("Failed to query a collateral position row [%s]", err)
		}

		fmt.Println("repoDealCloseSettlement: Collaterals Lender JSON", string(collLenderRow))
		if string(collLenderRow) != "" {
			// Read multiple collaterals
			var carbitrary_json map[string]interface{}
			err = json.Unmarshal([]byte(collLenderRow), &carbitrary_json)
			if err != nil {
				fmt.Println("Error parsing JSON: ", err)
			}

			collateralData := carbitrary_json["Collateral"].([]interface{})
			fmt.Println("repoDealCloseSettlement: Collateral Data: %v", collateralData)

			for key1, value1 := range collateralData {
				var assetTranStruct AssetTransfers
				
				fmt.Printf("Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
				jsonByte, err := json.Marshal(value1)
				//jsonStr := convertArray(jsonByte)
				fmt.Println("repoDealCloseSettlement: Collateral Data JSON Str: %v", string(jsonByte))

				var collStruct CollateralDetails
				err = json.Unmarshal([]byte(jsonByte), &collStruct)
				if err != nil {
					fmt.Println("Error retriving Repo Collaterals information", err)
					fmt.Println("Error parsing collDetails JSON: ", err)
				}

				if collStruct.SubAccount != "" {
					
					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.BorrowerParticipantID
					assetTranStruct.FromAcct = collStruct.SubAccount
					assetTranStruct.ToParty = collStruct.LenderParticipantID
					assetTranStruct.ToAcct = collStruct.LenderParticipantAcct	
			
				} else {

					assetTranStruct.InstrumentID = collStruct.Instrument
					assetTranStruct.PositionQty = collStruct.Quantity
					assetTranStruct.FromParty = collStruct.BorrowerParticipantID
					assetTranStruct.FromAcct = collStruct.BorrowerParticipantAcct
					assetTranStruct.ToParty = collStruct.LenderParticipantID
					assetTranStruct.ToAcct = collStruct.LenderParticipantAcct

				}
					
				assetTranArray = append(assetTranArray, assetTranStruct)
			}
		}
	} 
	
	settInstJSON, err = instHandler.newSettInstructionEntry(stub, sysReference, tradeStruct.TradeType, assetTranArray, "CloseLeg Settlement", settlementType, "STOCK", repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error generating settlement instruction", err)
		return nil, nil, err
	}
	
	if tradeStruct.TradeType == "REPURCHASE" || tradeStruct.TradeType == "REVERSEREPURCHASE" {

		var cashTranStruct AssetTransfers
		var currentAccrAmount float64
		var currentCashAmount float64
		var finalBorrowerCashAmount float64
		currentAccrAmount, _ = strconv.ParseFloat(tradeStruct.AccruedInterest, 64)		
		currentCashAmount, _ = strconv.ParseFloat(tradeStruct.TotalCashAmount, 64)		
		finalBorrowerCashAmount = currentCashAmount + currentAccrAmount		
		finalBorrowerCashAmountStr := strconv.FormatFloat(finalBorrowerCashAmount, 'f', 2, 64) 

		cashTranStruct.InstrumentID = tradeStruct.SettleCurrency
		cashTranStruct.PositionQty = finalBorrowerCashAmountStr
		cashTranStruct.FromParty = colllender.ParticipantID
		cashTranStruct.FromAcct = colllender.TradingAccount
		cashTranStruct.ToParty = collborrower.ParticipantID
		cashTranStruct.ToAcct = collborrower.TradingAccount

		cashTranArray = append(cashTranArray, cashTranStruct)

		cashInstJSON, err := instHandler.newSettInstructionEntry(stub, sysReference, tradeStruct.TradeType, cashTranArray, "CloseLeg Settlement", settlementType, "CASH", repoChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error generating settlement instruction", err)
			return nil, nil, err
		}
		return []byte(settInstJSON), []byte(cashInstJSON), nil
	}

	return []byte(settInstJSON), nil, nil
}

func (t *repoHandler) interimInterestPaymentSettlement(stub shim.ChaincodeStubInterface, sysReference string, interimPayment string, settlementType string, repoChaincodeID string, repoChannelID string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: interimInterestPayment ")
	var err error
	var colllender ParticipantDetails
	var collborrower ParticipantDetails
	var cashTranArray []AssetTransfers
//	settlementType = "RepoInterestPayment"

	if sysReference == "" || interimPayment == "" {
		shim.Error("SysReference or interimPayment is not passed")
	}

	fmt.Println("Querying trade for paying interim interest payment  :", sysReference)

	tradeJSON, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Error querying trade for paying interim interest payment:", sysReference, err)
		return nil, err
	}

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeJSON), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing trade JSON: ", err)
	}

	partyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Party, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo party information", err)
	}

	cpartyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Counterparty, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo counterparty information", err)
	}

	var partyStruct ParticipantDetails
	err = json.Unmarshal([]byte(partyJSON), &partyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling party JSON: ", err)
	}
	
	if partyStruct.PartyType == "COLLLENDER" {
		colllender = partyStruct
	} else {
		collborrower = partyStruct
	}

	var cpartyStruct ParticipantDetails
	err = json.Unmarshal([]byte(cpartyJSON), &cpartyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling cpty JSON: ", err)
	}
	
	if cpartyStruct.PartyType == "COLLLENDER" {
		colllender = cpartyStruct
	} else {
		collborrower = cpartyStruct
	}
	
	fmt.Println("Paying interim interest on the trade :", sysReference)
	err = tHandler.interestPayment(stub, tradeStruct, interimPayment)
	if err != nil {
		fmt.Println("Error paying interest on the trade :", sysReference, err)
		return nil, err
	}

	var cashTranStruct AssetTransfers
	cashTranStruct.InstrumentID = tradeStruct.SettleCurrency
	cashTranStruct.PositionQty = interimPayment			
	cashTranStruct.FromParty = colllender.ParticipantID
	cashTranStruct.FromAcct = colllender.TradingAccount
	cashTranStruct.ToParty = collborrower.ParticipantID
	cashTranStruct.ToAcct = collborrower.TradingAccount
		
	cashTranArray = append(cashTranArray, cashTranStruct)

	cashInstJSON, err := instHandler.newSettInstructionEntry(stub, sysReference, tradeStruct.TradeType, cashTranArray, "Interim Payment Settlement", settlementType, "CASH", repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error generating settlement instruction", err)
		return nil, err
	}
	
	return []byte(cashInstJSON), nil
}

func (t *repoHandler) initiateCashAdjustmentSettlement(stub shim.ChaincodeStubInterface, sysReference string, cashPayment string, indicator string, settlementType string, repoChaincodeID string, repoChannelID string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: initiateCashAdjustmentSettlement ")
	var err error
	var colllender ParticipantDetails
	var collborrower ParticipantDetails
	var cashTranArray []AssetTransfers
	//var settlementType string
	//settlementType = "RepoCashAdjustment"
	
	if sysReference == "" || cashPayment == "" {
		shim.Error("SysReference or cashPayment is not passed")
	}

	fmt.Println("Querying trade for paying Cash adjustment payment  :", sysReference)

	tradeJSON, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Error querying trade for paying cash adjustment payment:", sysReference, err)
		return nil, err
	}

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeJSON), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing trade JSON: ", err)
	}

	partyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Party, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo party information", err)
	}

	cpartyJSON, err := partHandler.queryParticipant(stub, tradeStruct.Counterparty, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo counterparty information", err)
	}

	var partyStruct ParticipantDetails
	err = json.Unmarshal([]byte(partyJSON), &partyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling party JSON: ", err)
	}
	
	if partyStruct.PartyType == "COLLLENDER" {
		colllender = partyStruct
	} else {
		collborrower = partyStruct
	}

	var cpartyStruct ParticipantDetails
	err = json.Unmarshal([]byte(cpartyJSON), &cpartyStruct)
	if err != nil {
		fmt.Println("Error unmarshalling cpty JSON: ", err)
	}
	
	if cpartyStruct.PartyType == "COLLLENDER" {
		colllender = cpartyStruct
	} else {
		collborrower = cpartyStruct
	} 

	fmt.Println("Paying cash adjustment on the trade :", sysReference)
	err = tHandler.cashAdjustmentPayment(stub, tradeStruct, cashPayment, indicator)
	if err != nil {
		fmt.Println("Error paying cash adjustment on the trade :", sysReference, err)
		return nil, err
	}

	var cashTranStruct AssetTransfers
	cashTranStruct.InstrumentID = tradeStruct.SettleCurrency
	cashTranStruct.PositionQty = cashPayment		

	if indicator == "DEBIT" {	
		cashTranStruct.FromParty = collborrower.ParticipantID
		cashTranStruct.FromAcct = collborrower.TradingAccount
		cashTranStruct.ToParty = colllender.ParticipantID
		cashTranStruct.ToAcct = colllender.TradingAccount				
	} else {
		cashTranStruct.FromParty = colllender.ParticipantID
		cashTranStruct.FromAcct = colllender.TradingAccount
		cashTranStruct.ToParty = collborrower.ParticipantID
		cashTranStruct.ToAcct = collborrower.TradingAccount
	}		

	cashTranArray = append(cashTranArray, cashTranStruct)

	cashInstJSON, err := instHandler.newSettInstructionEntry(stub, sysReference, tradeStruct.TradeType, cashTranArray, "Cash Adjustment Settlement", settlementType, "CASH", repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error generating settlement instruction", err)
		return nil, err
	}
	
	return []byte(cashInstJSON), nil
	
}

func (t *repoHandler) repoDealCollateralSub(stub shim.ChaincodeStubInterface, sysReference string, repoJSON string, settlementType string, repoChaincodeID string, repoChannelID string ) ([]byte, error) {
	
	fmt.Println("###### RepoDealCC: function: repoDealCollateralSub ")
	var err error
	var newCollateralData []interface{}	
	var assetTranArray []AssetTransfers	
	
	fmt.Println("repoDealCollateralSub: Coll Subs Input String [%s]", repoJSON)

	// -------------- Retriving new coll subs-----------------
	var arbitrary_json map[string]interface{}
	err = json.Unmarshal([]byte(repoJSON), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}
	jsonByte, err := json.Marshal(arbitrary_json["Trade"])
	jsonStr := convertArray(jsonByte)

	fmt.Println("repoDealCollateralSub: Trade Data: %s", string(jsonStr))
	var tradeStruct TradeDetails

	err = json.Unmarshal([]byte(jsonStr), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing JSON [%v]", err)
	}

	jsonRepoStr, err := t.getRepoDealInformation(stub, sysReference)
	if err != nil {
		fmt.Println("Error retriving Repo deal information", err)
		return nil, err
	}

	if string(jsonRepoStr) == "" {
		fmt.Println("Repo Deal is not found!")
		return nil, nil
	}

	var repo_arbitrary_json map[string]interface{}
	//var repoInputStruct RepoDeal
	err = json.Unmarshal([]byte(jsonRepoStr), &repo_arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}

	fmt.Println("repoDealCollateralSub: Trade Data: %v", repo_arbitrary_json["Trade"])
	tradeDetails, err := json.Marshal(repo_arbitrary_json["Trade"])
	tradeDetailsStr := convertArray(tradeDetails)
	//var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeDetailsStr), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}
	fmt.Println("repoDealCollateralSub: Trade Data: %v", tradeStruct.ProcessingSystemReference)
/*
	partyDetails, err := json.Marshal(repo_arbitrary_json["Party"])
	partyDetailsStr := convertArray(partyDetails)
	var partyStruct ParticipantDetails
	err = json.Unmarshal([]byte(partyDetailsStr), &partyStruct)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}
	fmt.Println("Party Data: %v", partyStruct.ParticipantID)

	if partyStruct.PartyType == "COLLLENDER" {
		colllender = partyStruct
	} else {
		collborrower = partyStruct
	}
	
	cpartyDetails, err := json.Marshal(repo_arbitrary_json["Counterparty"])
	cpartyDetailsStr := convertArray(cpartyDetails)
	var cpartyStruct ParticipantDetails
	err = json.Unmarshal([]byte(cpartyDetailsStr), &cpartyStruct)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}
	fmt.Println("Counterparty Data: %v", cpartyStruct.ParticipantID)

	if cpartyStruct.PartyType == "COLLLENDER" {
		colllender = cpartyStruct
	} else {
		collborrower = cpartyStruct
	} */

	//Capture New Multiple Collaterals
	newCollateralData = arbitrary_json["Collateral"].([]interface{})
	fmt.Println("repoDealCollateralSub: New Collateral Data:: ", newCollateralData)
	
	for key1, value1 := range newCollateralData {
		var assetTranStruct AssetTransfers
		
		fmt.Printf("repoDealCollateralSub: New Collateral Position Data index:%s  value1:%v  kind:%s  type:%s\n", key1, value1, reflect.TypeOf(value1).Kind(), reflect.TypeOf(value1))
		newjsonByte, err := json.Marshal(value1)
		newjsonStr := convertArray(newjsonByte)
		fmt.Println("repoDealCollateralSub: New Collateral Data JSON Str: %v", newjsonStr)

		var newCollStruct CollateralDetails
		err = json.Unmarshal([]byte(newjsonByte), &newCollStruct)
		if err != nil {
			fmt.Println("Error retriving Repo Collaterals information", err)
			fmt.Println("Error parsing collDetails JSON: ", err)
		}

		//New Collateral Added
		if newCollStruct.EditFlag == "A" {
			newCollStruct.ProcessingSystemReference = sysReference
			fmt.Println("repoDealCollateralSub: ########## Capturing new Collateral [%s]", newCollStruct.Instrument)
			err = collHandler.newCollateralPosition(stub, newCollStruct)
			if err != nil {
				fmt.Println("Collateral Capture failed for: %v : %v", arbitrary_json["Collateral"].(string), err)
				return nil, err
			}	

			if newCollStruct.SubAccount != "" {
				
				assetTranStruct.InstrumentID = newCollStruct.Instrument
				assetTranStruct.PositionQty = newCollStruct.Quantity				
				assetTranStruct.FromParty = newCollStruct.LenderParticipantID
				assetTranStruct.FromAcct = newCollStruct.LenderParticipantAcct	
				assetTranStruct.ToParty = newCollStruct.BorrowerParticipantID
				assetTranStruct.ToAcct = newCollStruct.SubAccount
			} else {

				assetTranStruct.InstrumentID = newCollStruct.Instrument
				assetTranStruct.PositionQty = newCollStruct.Quantity
				assetTranStruct.FromParty = newCollStruct.LenderParticipantID
				assetTranStruct.FromAcct = newCollStruct.LenderParticipantAcct
				assetTranStruct.ToParty = newCollStruct.BorrowerParticipantID
				assetTranStruct.ToAcct = newCollStruct.BorrowerParticipantAcct
			}

			assetTranArray = append(assetTranArray, assetTranStruct)

		} else if newCollStruct.EditFlag == "D" {  //Deleted Existing Collateral

			fmt.Println("repoDealCollateralSub: ########## Deactivating Collateral [%s]", newCollStruct.Instrument)

			fmt.Println("Reading existing collateral List by instrument", newCollStruct.BorrowerParticipantID, sysReference, newCollStruct.Instrument)
			excollJSONBytes, err := collHandler.queryCollateralPositionByInstrument(stub, newCollStruct.BorrowerParticipantID, sysReference, newCollStruct.Instrument)
			if err != nil {
				fmt.Println("Error retriving Repo Collaterals information", err)
			}

			var exCollStruct CollateralDetails
			err = json.Unmarshal([]byte(excollJSONBytes), &exCollStruct)
			if err != nil {
				fmt.Println("Error retriving Repo Collaterals information", err)
				fmt.Println("Error parsing collDetails JSON: ", err)
			}
		
			err = collHandler.deactivateCollateralPosition(stub, exCollStruct)
			if err != nil {
				fmt.Println("Error deactivating Repo Collaterals", err)
			}

			if exCollStruct.SubAccount != "" {
				
				assetTranStruct.InstrumentID = exCollStruct.Instrument
				assetTranStruct.PositionQty = exCollStruct.Quantity				
				assetTranStruct.FromParty = exCollStruct.BorrowerParticipantID
				assetTranStruct.FromAcct = exCollStruct.SubAccount
				assetTranStruct.ToParty = exCollStruct.LenderParticipantID
				assetTranStruct.ToAcct = exCollStruct.LenderParticipantAcct	

			} else {

				assetTranStruct.InstrumentID = exCollStruct.Instrument
				assetTranStruct.PositionQty = exCollStruct.Quantity
				assetTranStruct.FromParty = exCollStruct.BorrowerParticipantID
				assetTranStruct.FromAcct = exCollStruct.BorrowerParticipantAcct
				assetTranStruct.ToParty = exCollStruct.LenderParticipantID
				assetTranStruct.ToAcct = exCollStruct.LenderParticipantAcct

			}
					
			assetTranArray = append(assetTranArray, assetTranStruct)

		} else if newCollStruct.EditFlag == "X" {  //Existing Collateral unchanged
			fmt.Println("repoDealCollateralSub: Collateral unchanged ")
		}
	}

	settInstJSON, err := instHandler.newSettInstructionEntry(stub, sysReference, tradeStruct.TradeType, assetTranArray, "CloseLeg Settlement", settlementType, "STOCK", repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error generating settlement instruction", err)
		return nil, err
	}

	return []byte(settInstJSON), nil
}