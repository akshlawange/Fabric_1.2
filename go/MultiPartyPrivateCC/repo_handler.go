package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//repoHandler provides APIs used to perform operations on CC's KV store
type repoHandler struct {
}

// NewParticipantsHandler create a new participants
func NewRepoHandler() *repoHandler {
	return &repoHandler{}
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

func (t *repoHandler) getRepoDeal(stub shim.ChaincodeStubInterface, transactionRef string, repoChaincodeID string, repoChannelID string) ([]byte, error) {

	fmt.Println("###### MultiPartyChaincode: function: getRepoDeal ")

	if string(transactionRef) != "" && repoChaincodeID != "" && repoChannelID != "" {
		f := "query"
		invokeArgs := util.ToChaincodeArgs(f, "getRepoDealInformationWOAC", string(transactionRef))
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

func (t *repoHandler) deployRepoDeal(stub shim.ChaincodeStubInterface, transactionRef string, repoJSON string, notificationType string, multiPartyStatus string, repoChaincodeID string, repoChannelID string) error {

	fmt.Println("###### MultiPartyChaincode: function: invoke ")

	if transactionRef != "" && repoJSON != "" && repoChaincodeID != "" && repoChannelID != "" && notificationType != "" && multiPartyStatus != "" {

		f := "invoke"
		invokeArgs := util.ToChaincodeArgs(f, "repoDealApproval", transactionRef, string(repoJSON), notificationType, multiPartyStatus)
		response := stub.InvokeChaincode(repoChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke Repo chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}

/*
func (t *repoHandler) deployCollateralSub(stub shim.ChaincodeStubInterface, transactionRef string, repoJSON string, repoChaincodeID string, collSubChaincodeID string, repoChannelID string) error {

	fmt.Println("###### MultiPartyChaincode: function: deployCollateralSub ")

	if repoJSON != "" && repoChaincodeID != "" && repoChannelID != "" && collSubChaincodeID != "" {

		f := "invoke"
		invokeArgs := util.ToChaincodeArgs(f, "initiateRepoDealCollateralSub", transactionRef, string(repoJSON), repoChaincodeID, repoChannelID)
		response := stub.InvokeChaincode(collSubChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke Repo chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}

func (t *repoHandler) deployRepoClose(stub shim.ChaincodeStubInterface, transactionRef string, repoChaincodeID string, repoChannelID string) error {

	fmt.Println("###### MultiPartyChaincode: function: deployRepoClose ")

	if transactionRef != "" && repoChaincodeID != "" && repoChannelID != "" {

		f := "invoke"
		invokeArgs := util.ToChaincodeArgs(f, "repoDealClose", transactionRef)
		response := stub.InvokeChaincode(repoChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke Repo chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}

func (t *repoHandler) deployRepoInterestPayment(stub shim.ChaincodeStubInterface, transactionRef string, payment string, repoChaincodeID string, repoChannelID string) error {

	fmt.Println("###### MultiPartyChaincode: function: deployRepoInterestPayment ")

	if transactionRef != "" && payment != "" && repoChaincodeID != "" && repoChannelID != "" {

		f := "invoke"
		invokeArgs := util.ToChaincodeArgs(f, "initiateInterimInterestPaymentSettlement", transactionRef, string(payment))
		response := stub.InvokeChaincode(repoChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke Repo chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
} */

/*
func (t *repoHandler) setRepoDealStatus(stub shim.ChaincodeStubInterface, transactionRef string, repoChaincodeID string, repoChannelID string) error {

	fmt.Println("###### MultiPartyChaincode: function: deployRepoInterestPayment ")

	if transactionRef != "" && payment != "" && repoChaincodeID != "" && repoChannelID != "" {

		f := "invoke"
		invokeArgs := util.ToChaincodeArgs(f, "initiateInterimInterestPaymentSettlement", transactionRef, string(payment))
		response := stub.InvokeChaincode(repoChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke Repo chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}*/
