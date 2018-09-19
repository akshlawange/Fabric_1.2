package main

import (
//	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	
)

//tokenHandler provides APIs used to perform operations on CC's KV store
type tokenHandler struct {
}

// NewTokenHandler create a new participants
func NewTokenHandler() *tokenHandler {
	return &tokenHandler{}
}

type AssetToken struct {
	ObjectType            string `json:"ObjectType,omitempty"`                //ASSETTOKEN
	AssetTokenIdentifier  string `json:"AssetTokenIdentifier,omitempty"`
	InstrumentID          string `json:"InstrumentID,omitempty"`			
	OwnerParty            string `json:"OwnerParty,omitempty"`
	OwnerAcct             string `json:"OwnerAcct,omitempty"`
	PositionQty           string `json:"PositionQty,omitempty"`	
	Currency              string `json:"Currency,omitempty"`
	TransferDateTime      string `json:"IssuanceDateTime,omitempty"`	
	ActiveInd             string `json:"ActiveInd,omitempty"`
	UpdatedByUser         string `json:"UpdatedByUser,omitempty"`
	LastUpdatedTimestamp  string `json:"LastUpdatedTimestamp,omitempty"`
	TimeValue             string `json:"TimeValue,omitempty"`
	Version               int    `json:"Version,omitempty"`
}

// newTokenIssuance adds the new tokens in the chaincode state
func (t *tokenHandler) tokenTranfer(stub shim.ChaincodeStubInterface, assetTokenJSON string, fromOwnerParty string, toOwnerParty string, toOwnerAcct string, quantity string, assetTokenChaincoodeID string, assetTokenChannelID string) error {

	fmt.Println("###### SettlementContract: function: tokenTranfer ")

	if string(assetTokenJSON) != "" && string(fromOwnerParty) != "" && string(toOwnerParty) != "" && string(toOwnerAcct) != "" && string(quantity) != "" && assetTokenChaincoodeID != "" && assetTokenChannelID != "" {
		f := "invoke"
		fromCollection := fromOwnerParty + "Wallet"
		toCollection := toOwnerParty + "Wallet"
		invokeArgs := util.ToChaincodeArgs(f, "transferAssetToken", string(fromCollection), string(toCollection), string(assetTokenJSON), string(toOwnerParty), string(toOwnerAcct), string(quantity))
		response := stub.InvokeChaincode(assetTokenChaincoodeID, invokeArgs, assetTokenChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke Asset Token chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("tokenTranfer: Invoke Asset Token chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}

func (t *tokenHandler) readAvailableTokenPosition(stub shim.ChaincodeStubInterface, InstrumentID string, OwnerParty string, OwnerAcct string, assetTokenChaincoodeID string, assetTokenChannelID string) ([]byte, error) {
	
	fmt.Println("###### SettlementContract: function: readAvailableTokenPosition ")
	
	f := "query"
	collection := OwnerParty + "Wallet"
	invokeArgs := util.ToChaincodeArgs(f, "getAssetTokenInformation", collection, InstrumentID, OwnerParty, OwnerAcct)
	response := stub.InvokeChaincode(assetTokenChaincoodeID, invokeArgs, assetTokenChannelID)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to query Asset Token chaincode for getAssetTokenInformation. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}
	
	fmt.Printf("Query Asset Token chaincode successful. Got response %s", string(response.Payload))
	
	return []byte(response.Payload), nil
}
