package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//settlementInstHandler provides APIs used to perform operations on CC's KV store
type settlementInstHandler struct {
}

// NewPostingHandler create a new postings
func NewSettInstHandler() *settlementInstHandler {
	return &settlementInstHandler{}
}

type SettlementInstruction struct {
	ObjectType            string `json:"ObjectType,omitempty"`        //SETTINSTRUCTION
	SysReference          string `json:"SysReference,omitempty"`       	
	TradeType             string `json:"TradeType,omitempty"`       	
	AssetType             string `json:"AssetType,omitempty"`         // STOCK, CASH
	AssetTransfers        []AssetTransfers `json:"AssetTransfers,omitempty"`     		
	TransferDateTime      string `json:"TransferDateTime,omitempty"`	
	SettlementType   	  string `json:"SettlementType,omitempty"`	  //RepoOpenLegSettlement, RepoCloseLegSettlement, RepoCollateralSubstitution, RepoInterestPayment, RepoCashAdjustment
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
                                                                               
// newSettInstructionEntry adds the posting record on the chaincode state table
func (t *settlementInstHandler) newSettInstructionEntry(stub shim.ChaincodeStubInterface, sysReference string, tradeType string, assetTranArray []AssetTransfers, reason string, settlementType string, assetType string, repoChaincodeID string, repoChannelID string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: newSettInstructionEntry ")

	var instStruct SettlementInstruction
	instStruct.SysReference = sysReference
	instStruct.AssetTransfers = assetTranArray	
	instStruct.SettlementType = settlementType
	instStruct.TradeType = tradeType
	instStruct.AssetType = assetType
	instStruct.SettlementStatus = "PENDING"		
	instStruct.Reason = reason
	instStruct.RepoChaincodeID = repoChaincodeID
	instStruct.RepoChannelID = repoChannelID
	instJSONBytes, _ := json.Marshal(instStruct)
	fmt.Println("newSettInstructionEntry: New Instruction: ", string(instJSONBytes))

	return instJSONBytes, nil
}
