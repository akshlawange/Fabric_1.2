package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//CollateralHandler provides APIs used to perform operations on CC's KV store
type repoHandler struct {
}

// NewCollateralHandler create a new reference to CertHandler
func NewRepoHandler() *repoHandler {
	return &repoHandler{}
}

func (t *repoHandler) readRepoContracts(stub shim.ChaincodeStubInterface, participantID string, repoChaincodeID string, repoChannelID string) ([]byte, error) {

	fmt.Println("###### ValuationCC: function: readRepoContracts ")

	//Querying queryCollaterals in Repo Chaincode
	f := "query"
	invokeArgs := util.ToChaincodeArgs(f, "queryCollaterals", participantID)
	response := stub.InvokeChaincode(repoChaincodeID, invokeArgs, repoChannelID)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to query Repo chaincode for queryCollaterals. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Query Repo chaincode successful. Got response %s", string(response.Payload))

	return response.Payload, nil
}

func (t *repoHandler) readPriceByInstrument(stub shim.ChaincodeStubInterface, instrumentID string, refChaincodeID string, globalChannelID string) ([]byte, error) {

	fmt.Println("###### ValuationCC: function: readPriceByInstrument ")

	//Querying getSecurityPrice in reference Chaincode
	f := "query"
	invokeArgs := util.ToChaincodeArgs(f, "getSecurityPrice", instrumentID)
	response := stub.InvokeChaincode(refChaincodeID, invokeArgs, globalChannelID)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to query ReferenceData chaincode for queryPriceByInstrument. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	fmt.Printf("Invoke Reference Data chaincode successful for queryPriceByInstrument. Got response %s", string(response.Payload))

	return response.Payload, nil
}

func (t *repoHandler) updateRepoCollateral(stub shim.ChaincodeStubInterface, collUpdateJSON string, repoChaincodeID string, repoChannelID string) error {

	fmt.Println("###### ValuationCC: function: readPriceByInstrument ")

	//Invoking newCollateralSubUpdate in Repo Chaincode
	f := "invoke"
	invokeArgs := util.ToChaincodeArgs(f, "newCollateralSubUpdate", collUpdateJSON, "AMEND")
	response := stub.InvokeChaincode(repoChaincodeID, invokeArgs, repoChannelID)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke Repo chaincode for newCollateralSubUpdate. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)
		return errors.New(errStr)
	}

	fmt.Printf("Invoke Repo chaincode successful for newCollateralSubUpdate. Got response %s", string(response.Payload))

	return nil
}
