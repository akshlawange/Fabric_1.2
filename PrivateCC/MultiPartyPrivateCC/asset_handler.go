package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//repoHandler provides APIs used to perform operations on CC's KV store
type assetHandler struct {
}

// NewParticipantsHandler create a new participants
func NewAssetHandler() *assetHandler {
	return &assetHandler{}
}

func (t *assetHandler) deployAssetOwnership(stub shim.ChaincodeStubInterface, assetJSON string, notificationType string, assetOwnershipChaincodeID string, channelID string) error {

	fmt.Println("###### MultiPartyChaincode: function: deployAssetOwnership")

	if  notificationType == "TokenIssuance" && assetJSON != "" && assetOwnershipChaincodeID != "" && channelID != ""  {

		f := "invokeInternal"
		invokeArgs := util.ToChaincodeArgs(f, "newAssetCreation", string(assetJSON))
		response := stub.InvokeChaincode(assetOwnershipChaincodeID, invokeArgs, channelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke AssetOwnership chaincode successful. Got response %s", string(response.Payload))
	}

	if  notificationType == "TokenRedemption" && assetJSON != "" && assetOwnershipChaincodeID != "" && channelID != "" {

		f := "invokeInternal"
		invokeArgs := util.ToChaincodeArgs(f, "initiateAssetWithdrawal", string(assetJSON))
		response := stub.InvokeChaincode(assetOwnershipChaincodeID, invokeArgs, channelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke AssetOwnership chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}
