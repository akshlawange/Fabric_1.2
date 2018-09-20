package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type ParticipantDetails struct {
	ObjectType                string `json:"ObjectType,omitempty"`
	ProcessingSystemReference string `json:"ProcessingSystemReference,omitempty"`
	ParticipantID             string `json:"ParticipantID,omitempty"`
	ClientParticipantID       string `json:"ClientParticipantID,omitempty"` //POPULATE IN CASE OF CUSTODIAN
	Book                      string `json:"Book,omitempty"`
	TradingAccount            string `json:"TradingAccount,omitempty"`
	ParentAccount             string `json:"ParentAccount,omitempty"`
	PartyType                 string `json:"PartyType,omitempty"` // COLLLENDER, COLLBORROWER, CUSTODIAN
	Entity                    string `json:"Entity,omitempty"`    //LEGAL ENTITY OF
	InternalTradeReference    string `json:"InternalTradeReference,omitempty"`
	UserID                    string `json:"UserID,omitempty"`
	LastUpdatedUser           string `json:"LastUpdatedUser,omitempty"`
	DateTime                  string `json:"DateTime,omitempty"`
	Version                   int    `json:"Version,string,omitempty"`
	ActiveInd                 string `json:"ActiveInd,omitempty"`
}

//ParticipantHandler provides APIs used to perform operations on CC's KV store
type participantHandler struct {
}

// NewParticipantsHandler create a new participants
func NewParticipantsHandler() *participantHandler {
	return &participantHandler{}
}

//var utilHandler = NewUtilityHandler()

// newParticipants adds the record row on the chaincode state table
func (t *participantHandler) newParticipants(stub shim.ChaincodeStubInterface, partStruct ParticipantDetails) error {

	fmt.Println("###### RepoDealCC: function: newParticipants ")

	partStruct.ObjectType = "ParticipantDetails"
	partStruct.Version = 1
	partStruct.ActiveInd = "A"
	partStruct.DateTime = time.Now().UTC().String()
	collection := "RepoDealCollection" 

	compositeKey, _ := stub.CreateCompositeKey(partStruct.ObjectType, []string{partStruct.ParticipantID, partStruct.ProcessingSystemReference, strconv.Itoa(partStruct.Version)})
	partJSONBytes, _ := json.Marshal(partStruct)

	err := stub.PutPrivateData(collection,compositeKey, partJSONBytes)
	if err != nil {
		return errors.New("Error in adding participant state")
	}

	return nil
}

// updateParticipants replaces the participant record row on the chaincode state table
func (t *participantHandler) updateParticipants(stub shim.ChaincodeStubInterface, partStruct ParticipantDetails) error {

	fmt.Println("###### RepoDealCC: function: updateParticipants ")
	//Get Current State
	partStruct.ObjectType = "ParticipantDetails"
	collection:= "RepoDealCollection"
	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(partStruct.ObjectType, []string{partStruct.ParticipantID, partStruct.ProcessingSystemReference, strconv.Itoa(partStruct.Version)})
	extJSONBytes, _ := stub.GetPrivateData(collection,compositeKey)

	var exPartyStruct ParticipantDetails
	if string(extJSONBytes) != "" {
		err := json.Unmarshal([]byte(extJSONBytes), &exPartyStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}

		exPartyStruct.ActiveInd = "N"
		exPartyStruct.DateTime = time.Now().UTC().String()
		extJSONBytes, _ = json.Marshal(exPartyStruct)

		err = stub.PutPrivateData(collection,compositeKey, extJSONBytes)
		if err != nil {
			return errors.New("Error in updating participant state")
		}
	}
	//Create new version and document
	partStruct.Version = partStruct.Version + 1
	partStruct.ActiveInd = "A"
	partStruct.DateTime = time.Now().UTC().String()
	collection := "RepoDealCollection"

	compositeKey, _ = stub.CreateCompositeKey(partStruct.ObjectType, []string{partStruct.ParticipantID, partStruct.ProcessingSystemReference, strconv.Itoa(partStruct.Version)})
	partJSONBytes, _ := json.Marshal(partStruct)
	err := stub.PutPrivateData(collection,compositeKey, partJSONBytes)
	if err != nil {
		return errors.New("Error in adding participant state")
	}

	return nil
}

// updateParticipants replaces the participant record row on the chaincode state table
func (t *participantHandler) deactivateParticipants(stub shim.ChaincodeStubInterface, partStruct ParticipantDetails) error {

	fmt.Println("###### RepoDealCC: function: deactivateParticipants ")
	//Get Current State
	partStruct.ObjectType = "ParticipantDetails"
	collection:= "RepoDealCollection"

	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(partStruct.ObjectType, []string{partStruct.ParticipantID, partStruct.ProcessingSystemReference, strconv.Itoa(partStruct.Version)})
	extJSONBytes, _ := stub.GetPrivateData(collection,compositeKey)

	var exPartyStruct ParticipantDetails
	if string(extJSONBytes) != "" {
		err := json.Unmarshal([]byte(extJSONBytes), &exPartyStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
		exPartyStruct.ActiveInd = "N"
		exPartyStruct.DateTime = time.Now().UTC().String()
		extJSONBytes, _ = json.Marshal(exPartyStruct)
		err = stub.PutPrivateData(collection,compositeKey, extJSONBytes)
		if err != nil {
			return errors.New("Error in deactivating participant state")
		}
	}
	return nil
}

func (t *participantHandler) newParticipantCapture(stub shim.ChaincodeStubInterface, partyStruct ParticipantDetails, partyType string, action string) error {

	fmt.Println("###### RepoDealCC: function: newParticipantCapture ")
	var err error
	if action == "NEW" {
		err = t.newParticipants(stub, partyStruct)
		if err != nil {
			fmt.Println("Error adding new participant [%v]", err)
		}

	} else if action == "AMEND" {
		err = t.updateParticipants(stub, partyStruct)
		if err != nil {
			fmt.Println("Error updating participant [%v]", err)
		}

	} else if action == "CANCEL" {
		err = t.deactivateParticipants(stub, partyStruct)
		if err != nil {
			fmt.Println("Error deactivating participant [%v]", err)
		}
	}

	return nil
}

// queryParticipant returns the record row matching a corresponding position on the chaincode state
func (t *participantHandler) queryParticipant(stub shim.ChaincodeStubInterface, participantID string, processingSystemReference string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryParticipant ")
	if participantID != "" && processingSystemReference != "" {
		var attributes []string
		attributes = append(attributes, participantID)
		attributes = append(attributes, processingSystemReference)
		partJSONBytes, err := utilHandler.readMultiParticipantsJSON(stub, "ParticipantDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving participant")
		}
		return partJSONBytes, nil
	}
	return nil, nil
}
