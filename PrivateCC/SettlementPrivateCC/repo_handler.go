package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	
)

//repoHandler provides APIs used to perform operations on CC's KV store
type repoHandler struct {
}

// NewRepoHandler to update repo status
func NewRepoHandler() *repoHandler {
	return &repoHandler{}
}

type SettlementAck struct {
	SysReference          string `json:"SysReference,omitempty"`			
	RepoStatus            string `json:"RepoStatus,omitempty"`	
	RepoChaincodeID       string `json:"RepoChaincodeID,omitempty"`
	RepoChannelID         string `json:"RepoChannelID,omitempty"`
}

func (t *repoHandler) repoSettlementAcknowledge(stub shim.ChaincodeStubInterface, sysReference string, settlementType string, repoChaincodeID string, repoChannelID string) ([]byte, error) {
	
	var err error
	fmt.Println("###### SettlementCC: function: repoSettlementAcknowledge ")
	/*
	if sysReference == "" || settlementType == "" {
		return nil, err
	}*/
	
	fmt.Println("repoSettlementAcknowledge: Querying trade for settlement status ACK  :", sysReference)
	
	repoStatus, err := t.getRepoDealStatus(stub, sysReference, repoChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Error querying trade repoStatus:", sysReference, err)
		return nil, err
	}
	fmt.Println("repoSettlementAcknowledge : Querying trade Repo status  :", string(repoStatus))
	
	repoStatusStr := string(repoStatus)
	if settlementType == "RepoOpenLegSettlement" && repoStatusStr == "OPENLEGPEND" {
		repoStatusStr = "OPENLEGSETTLED"
	} else if settlementType == "RepoCloseLegSettlement" && repoStatusStr == "CLOSELEGPEND" {
		repoStatusStr = "CLOSELEGSETTLED"
	} else if settlementType == "RepoCollateralSubstitution" && repoStatusStr == "COLLSUBPEND" {
		repoStatusStr = "OPENLEGSETTLED"
	} else if settlementType == "RepoInterestPayment" && repoStatusStr == "INTPAYMENTPEND" {
		repoStatusStr = "OPENLEGSETTLED"
	} else if settlementType == "RepoCashAdjustment" && repoStatusStr == "CASHADJPEND" {
		repoStatusStr = "OPENLEGSETTLED"
	}

	fmt.Println("repoSettlementAcknowledge: New Repo status  :", repoStatusStr)
	var settAckStruct SettlementAck
	settAckStruct.SysReference = sysReference
	settAckStruct.RepoStatus = repoStatusStr
	settAckStruct.RepoChaincodeID = repoChaincodeID
	settAckStruct.RepoChannelID = repoChannelID
	settAckJSON, _ := json.Marshal(settAckStruct)

	fmt.Println("repoSettlementAcknowledge: Settlement ACK JSON :", string(settAckJSON))
	
	return []byte(settAckJSON), nil
}
	

func (t *repoHandler) getRepoDealStatus(stub shim.ChaincodeStubInterface, transactionRef string, repoChaincodeID string, repoChannelID string) ([]byte, error) {
		
	fmt.Println("###### MultiPartyChaincode: function: getRepoDealStatus ")
	
	if string(transactionRef) != "" && repoChaincodeID != "" && repoChannelID != "" {
		f := "query"
		invokeArgs := util.ToChaincodeArgs(f, "getRepoDealRepoStatus", string(transactionRef))
		response := stub.InvokeChaincode(repoChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", string(response.Payload))
			return nil, errors.New(errStr)
		}
		
		fmt.Printf("Query Repo chaincode successful. Got response %s", string(response.Payload))
		return []byte(response.Payload), nil
	}
		
	return nil, nil
}
		
			