package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SettlementInstruction struct {
	ObjectType            string `json:"ObjectType,omitempty"`        //SETTINSTRUCTION
	SysReference          string `json:"SysReference,omitempty"`       
	AssetTransfers        []AssetTransfers `json:"AssetTransfers,omitempty"`     	
	TransferDateTime      string `json:"TransferDateTime,omitempty"`	
	SettlementType   	  string `json:"SettlementStatus,omitempty"`	  //RepoOpenLegSettlement, RepoCloseLegSettlement, RepoCollateralSubstitution, RepoInterestPayment, RepoCashAdjustment
	SettlementStatus	  string `json:"SettlementStatus,omitempty"`	  //SETTLED, PENDING, FAILED
	Reason                string `json:"Reason,omitempty"`	  
	RepoChaincodeID       string `json:"RepoChaincodeID,omitempty"`
	RepoChannelID         string `json:"RepoChannelID,omitempty"`	    
	ActiveInd             string `json:"ActiveInd,omitempty"`
	UpdatedByUser         string `json:"UpdatedByUser,omitempty"`
	LastUpdatedTimestamp  string `json:"LastUpdatedTimestamp,omitempty"`
	Version               int    `json:"Version,omitempty"`
}

type AssetTransfers struct {
	InstrumentID          string `json:"InstrumentID,omitempty"`			
	PositionQty           string `json:"PositionQty,omitempty"`	
	FromParty             string `json:"FromParty,omitempty"`
	FromAcct              string `json:"FromAcct,omitempty"`
	ToParty               string `json:"ToParty,omitempty"`
	ToAcct                string `json:"ToAcct,omitempty"`
}

//RepositoryHandler provides APIs used to perform operations on CC's KV store
type repositoryHandler struct {
}

// NewRepositoryHandler create a new reference to CertHandler
func NewRepositoryHandler() *repositoryHandler {
	return &repositoryHandler{}
}

var utilHandler = NewUtilityHandler()

// newSettlementInstruction adds the record row on the chaincode state
func (t *repositoryHandler) newSettlementInstruction(stub shim.ChaincodeStubInterface, settInstruction SettlementInstruction) error {
	fmt.Println("###### SettlementContract: function: newSettlementInstruction ")

	fmt.Println("newSettlementInstruction: insert settlement instruction= %v", settInstruction)

	settInstruction.ObjectType = "SETTINSTRUCTION"
	settInstruction.SettlementStatus = "PENDING"
	settInstruction.ActiveInd = "A"	
	settInstruction.Version = 1
	settInstruction.LastUpdatedTimestamp = time.Now().UTC().String()
	compositeKey, _ := stub.CreateCompositeKey(settInstruction.ObjectType, []string{settInstruction.SysReference, strconv.Itoa(settInstruction.Version)})
	assetJSONBytes, _ := json.Marshal(settInstruction)
	fmt.Println("New settlement instruction Creation: ", string(compositeKey), string(assetJSONBytes))

	stub.PutState(compositeKey, assetJSONBytes)
	return nil
}


// updateSettlementInstruction replaces the position record row on the chaincode state
func (t *repositoryHandler) updateSettlementInstruction(stub shim.ChaincodeStubInterface, settInstruction SettlementInstruction) error {
	
		fmt.Println("###### SettlementContract: function: updateSettlementInstruction ")
	
		fmt.Println("updateSettlementInstruction: update asset= %v", settInstruction)
	
		settInstruction.ObjectType = "SETTINSTRUCTION"
		compositeKey, _ := stub.CreateCompositeKey(settInstruction.ObjectType, []string{settInstruction.SysReference, strconv.Itoa(settInstruction.Version)})
		assetJSONBytes, _ := stub.GetState(compositeKey)
	
		fmt.Println("Existing sett instruction : ", string(compositeKey), string(assetJSONBytes))
		var exSettInstStruct SettlementInstruction
		if string(assetJSONBytes) != "" {
			err := json.Unmarshal([]byte(assetJSONBytes), &exSettInstStruct)
			if err != nil {
				fmt.Println("Error parsing sett intruction JSON [%v]", err)
				return err
			}
	
			exSettInstStruct.ActiveInd = "N"
			exSettInstStruct.LastUpdatedTimestamp = time.Now().UTC().String()
			assetJSONBytes, _ = json.Marshal(exSettInstStruct)
			fmt.Println("Existing sett instruction : ", string(compositeKey), string(assetJSONBytes))
	
			err = stub.PutState(compositeKey, assetJSONBytes)
			if err != nil {
				return errors.New("Error in updating sett instruction state")
			}
		}
	
		// Create a new Version and document
		settInstruction.ActiveInd = "A"
		settInstruction.Version = settInstruction.Version + 1
		settInstruction.LastUpdatedTimestamp = time.Now().UTC().String()
		compositeKey, _ = stub.CreateCompositeKey(settInstruction.ObjectType, []string{settInstruction.SysReference, strconv.Itoa(settInstruction.Version)})
		assetJSONBytes, _ = json.Marshal(settInstruction)
		fmt.Println("New sett instruction : ", string(compositeKey), string(assetJSONBytes))
	
		err := stub.PutState(compositeKey, assetJSONBytes)
		if err != nil {
			return errors.New("Error in adding sett instruction state")
		}
	
		return nil
	}


// queryAsset returns the asset for corresponding position on the chaincode state
func (t *repositoryHandler) querySettlementInstruction(stub shim.ChaincodeStubInterface, SysReference string) ([]byte, error) {
	fmt.Println("###### SettlementContract: function: querySettlementInstruction ")

	fmt.Println("querySettlementInstruction: Querying asset token")
	if SysReference != "" {

		var attributes []string
		attributes = append(attributes, SysReference)
		
		assetJSONBytes, err := utilHandler.readMultiJSON(stub, "SETTINSTRUCTION", attributes)
		if err != nil {
			fmt.Println("querySettlementInstruction: Error querying position", err)
			return []byte(""), err
		}
		
		return assetJSONBytes, nil
	}

	return []byte(""), errors.New("ERROR: Not enough parameters are passed !")
}
