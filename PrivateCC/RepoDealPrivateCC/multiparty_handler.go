package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//multipartyHandler provides APIs used to perform operations on CC's KV store
type multipartyHandler struct {
}

// NewMultipartyHandler create a new participants
func NewMultipartyHandler() *multipartyHandler {
	return &multipartyHandler{}
}

type NotificationsDetails struct {
	ObjectType         string `json:"ObjectType,omitempty"`
	ParticipantID      string `json:"ParticipantID,omitempty"`
	TransactionRef     string `json:"TransactionRef,omitempty"`
	NotificationType   string `json:"NotificationType,omitempty"` //New, Amend, Cancel, Collateral Substitution, Interest Payment, Cash Adjustment
	NotificationID     int    `json:"NotificationID,omitempty"`
	NotificationDesc   string `json:"NotificationDesc,omitempty"`
	NotificationData   string `json:"NotificationData,omitempty"`
	NotificationStatus string `json:"NotificationStatus,omitempty"` //Pending, Approved, Rejected
	Comment            string `json:"Comment,omitempty"`
	ActionByUser       string `json:"ActionByUser,omitempty"`
	DateTime           string `json:"DateTime,omitempty"`
	ActiveInd          string `json:"ActiveInd,omitempty"`
}

type MultiParty struct {
	ObjectType     string `json:"ObjectType,omitempty"`
	TransactionRef string `json:"TransactionRef,omitempty"`
	EventType      string `json:"EventType,omitempty"`   //New,  Amend, Cancel, Collateral Substitution, Interest Payment, Close
	EventStatus    string `json:"EventStatus,omitempty"` //Pending, Approved, Rejected
	MinCount       int    `json:"MinCount,omitempty"`
	PendingCount   int    `json:"PendingCount,omitempty"`
	ApproveCount   int    `json:"ApproveCount,omitempty"`
	RejectCount    int    `json:"RejectCount,omitempty"`
	DateTime       string `json:"DateTime,omitempty"`
	ActionByUser   string `json:"ActionByUser,omitempty"`
	ActiveInd      string `json:"ActiveInd,omitempty"`
}

// newNotificationEntry adds the agreement record on the chaincode state table
func (t *multipartyHandler) newNotificationEntry(stub shim.ChaincodeStubInterface, newNotification NotificationsDetails, mpChaincodeID string, repoChannelID string) error {

	var err error
	fmt.Println("###### RepoDealCC: function: newNotificationEntry ")
	jsonByte, err := json.Marshal(newNotification)
	if err != nil {
		fmt.Println("Error during generating notification JSON")
		return err
	}

	if string(jsonByte) != "" && mpChaincodeID != "" && repoChannelID != "" {
		f := "invoke"
		invokeArgs := util.ToChaincodeArgs(f, "newNotificationCapture", string(jsonByte))
		response := stub.InvokeChaincode(mpChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke Repo chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}

// newMultiParty adds the MultiParty record on the chaincode state table
func (t *multipartyHandler) newMultiParty(stub shim.ChaincodeStubInterface, mpStruct MultiParty, mpChaincodeID string, repoChannelID string) error {

	var err error
	fmt.Println("###### RepoDealCC: function: newMultiParty ")
	mpJSON, err := json.Marshal(mpStruct)
	if err != nil {
		fmt.Println("Error during generating notification JSON")
		return err
	}

	if string(mpJSON) != "" && mpChaincodeID != "" && repoChannelID != "" {
		f := "invoke"
		invokeArgs := util.ToChaincodeArgs(f, "newMultiParty", string(mpJSON))
		response := stub.InvokeChaincode(mpChaincodeID, invokeArgs, repoChannelID)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return errors.New(errStr)
		}

		fmt.Printf("Invoke MultiParty chaincode successful. Got response %s", string(response.Payload))
	}

	return nil
}

func (t *multipartyHandler) newRepoDealNotification(stub shim.ChaincodeStubInterface, repoJSON string, tranType string, sysReference string, mpChaincodeID string, repoChannelID string) error {

	var err error
	fmt.Println("###### RepoDealChaincode: function: newRepoDealNotification ")

	var arbitrary_json map[string]interface{}
	err = json.Unmarshal([]byte(repoJSON), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}

	fmt.Println("Trade Data: %v", arbitrary_json["Trade"])
	jsonByte, err := json.Marshal(arbitrary_json["Trade"])
	jsonStr := convertArray(jsonByte)
	fmt.Println("Trade Data: %s", string(jsonStr))

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(jsonStr), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing JSON [%v]", err)
	}

	var notStruct NotificationsDetails
	notStruct.NotificationType = tranType
	notStruct.NotificationStatus = "Pending"
	notStruct.NotificationData = repoJSON
	notStruct.NotificationDesc = "Repo Deal Approval"

	if sysReference != "" {
		notStruct.TransactionRef = sysReference
	} else {
		notStruct.TransactionRef = tradeStruct.ProcessingSystemReference
	}

	if tradeStruct.TranOriginatorParty == tradeStruct.Party {
		notStruct.ParticipantID = tradeStruct.Counterparty

	} else if tradeStruct.TranOriginatorParty == tradeStruct.Counterparty {
		notStruct.ParticipantID = tradeStruct.Party

	}

	//Notification for Party
	err = t.newNotificationEntry(stub, notStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("MPCHAINID2::",mpChaincodeID,"\nRChannelID2::",repoChannelID)
		fmt.Println("Repo New Notification Entry failed", err)
		return err
	}

	var mpStruct MultiParty
	mpStruct.EventType = tranType
	mpStruct.MinCount = 1
	mpStruct.PendingCount = 1
	if sysReference != "" {
		mpStruct.TransactionRef = sysReference
	} else {
		mpStruct.TransactionRef = tradeStruct.ProcessingSystemReference
	}
	mpStruct.ActionByUser = "RepoChaincode"

	err = t.newMultiParty(stub, mpStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Repo New MultiParty Entry failed", err)
		return err
	}

	return nil
}

func (t *multipartyHandler) newSettlementNotification(stub shim.ChaincodeStubInterface, notificationData string, tranType string, sysReference string, mpChaincodeID string, repoChannelID string) error {

	var err error
	fmt.Println("###### RepoDealChaincode: function: newSettlementNotification ")

	tradeJSON, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Error querying trade", err)
	}

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(tradeJSON), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing JSON [%v]", err)
	}

	var notStruct NotificationsDetails
	notStruct.NotificationType = tranType
	notStruct.NotificationStatus = ""
	notStruct.NotificationData = notificationData
	notStruct.NotificationDesc = "Settlement Confirmation"

	if sysReference != "" {
		notStruct.TransactionRef = sysReference
	} else {
		notStruct.TransactionRef = tradeStruct.ProcessingSystemReference
	}

	notStruct.ParticipantID = tradeStruct.Party

	//Notification for Party
	err = t.newNotificationEntry(stub, notStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Settlement Confirmation Notification Entry failed", err)
		return err
	}

	notStruct.ParticipantID = tradeStruct.Counterparty

	//Notification for Counterparty
	err = t.newNotificationEntry(stub, notStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Settlement Confirmation Notification Entry failed", err)
		return err
	}

	return nil
}

func (t *multipartyHandler) newRepoCloseNotification(stub shim.ChaincodeStubInterface, sysReference string, repoJSON string, tranType string, tranOriginatorParty string, mpChaincodeID string, repoChannelID string) error {

	fmt.Println("###### RepoDealChaincode: function: newRepoCloseNotification ")
	
	var notStruct NotificationsDetails
	notStruct.NotificationType = tranType
	notStruct.NotificationStatus = "Pending"
	notStruct.NotificationData = repoJSON
	notStruct.NotificationDesc = "Deal Close Approval"
	notStruct.TransactionRef = sysReference

	tradeRow, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Failed to query a position row [%s]", err)
	}
	fmt.Println("Trade Information:", string(tradeRow))
	var tradeStruct TradeDetails

	if string(tradeRow) != "" {
		err = json.Unmarshal([]byte(tradeRow), &tradeStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
	}

	if tradeStruct.Party == tranOriginatorParty {
		notStruct.ParticipantID = tradeStruct.Counterparty

	} else if tradeStruct.Counterparty == tranOriginatorParty {
		notStruct.ParticipantID = tradeStruct.Party

	}

	//Notification for Party
	err = t.newNotificationEntry(stub, notStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Repo Deal close Notification Entry failed", err)
		return err
	}

	var mpStruct MultiParty
	mpStruct.EventType = tranType
	mpStruct.MinCount = 1
	mpStruct.PendingCount = 1
	mpStruct.TransactionRef = sysReference
	mpStruct.ActionByUser = "RepoChaincode"

	err = t.newMultiParty(stub, mpStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Repo Deal close MultiParty Entry failed", err)
		return err
	}

	return nil
}

func (t *multipartyHandler) newInterestPaymentNotification(stub shim.ChaincodeStubInterface, sysReference string, payment string, party string, counterparty string, tranType string, tranOriginatorParty string, mpChaincodeID string, repoChannelID string) error {

	fmt.Println("###### RepoDealChaincode: function: newInterestPaymentNotification ")

	//tranType = "InterestPayment"
	notificationString := "{\"transactionRef\":\"" + sysReference + "\",\"payment\":\"" + payment + "\",\"Party\":\"" + party + "\",\"Counterparty\":\"" + counterparty + "\",\"TranOriginatorParty\":\"" + tranOriginatorParty + "\"}"

	var notStruct NotificationsDetails
	notStruct.NotificationType = tranType
	notStruct.NotificationStatus = "Pending"
	notStruct.NotificationData = notificationString
	notStruct.NotificationDesc = "Interim Interest Payment Approval"
	notStruct.TransactionRef = sysReference

	tradeRow, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Failed to query a position row [%s]", err)
	}
	fmt.Println("Trade Information:", string(tradeRow))
	var tradeStruct TradeDetails

	if string(tradeRow) != "" {
		err = json.Unmarshal([]byte(tradeRow), &tradeStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
	}

	if tradeStruct.Party == tranOriginatorParty {
		notStruct.ParticipantID = tradeStruct.Counterparty

	} else if tradeStruct.Counterparty == tranOriginatorParty {
		notStruct.ParticipantID = tradeStruct.Party

	}

	//Notification for Party
	err = t.newNotificationEntry(stub, notStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Interest Payment Notification Entry failed", err)
		return err
	}

	var mpStruct MultiParty
	mpStruct.EventType = tranType
	mpStruct.MinCount = 1
	mpStruct.PendingCount = 1
	mpStruct.TransactionRef = sysReference
	mpStruct.ActionByUser = "RepoChaincode"

	err = t.newMultiParty(stub, mpStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Interest Payment MultiParty Entry failed", err)
		return err
	}

	return nil
}

func (t *multipartyHandler) newCashAdjustmentNotification(stub shim.ChaincodeStubInterface, sysReference string, payment string, indicator string, party string, counterparty string, tranType string, tranOriginatorParty string, mpChaincodeID string, repoChannelID string) error {

	fmt.Println("###### RepoDealChaincode: function: newCashAdjustmentNotification ")

	//tranType = "InterestPayment"
	notificationString := "{\"transactionRef\":\"" + sysReference + "\",\"payment\":\"" + payment + "\",\"indicator\":\"" + indicator + "\",\"Party\":\"" + party + "\",\"Counterparty\":\"" + counterparty + "\",\"TranOriginatorParty\":\"" + tranOriginatorParty + "\"}"
	var notStruct NotificationsDetails
	notStruct.NotificationType = tranType
	notStruct.NotificationStatus = "Pending"
	notStruct.NotificationData = notificationString
	notStruct.NotificationDesc = "Cash Adjustment Approval"
	notStruct.TransactionRef = sysReference

	tradeRow, err := tHandler.queryActiveTrade(stub, sysReference)
	if err != nil {
		fmt.Println("Failed to query a position row [%s]", err)
	}
	fmt.Println("Trade Information:", string(tradeRow))
	var tradeStruct TradeDetails

	if string(tradeRow) != "" {
		err = json.Unmarshal([]byte(tradeRow), &tradeStruct)
		if err != nil {
			fmt.Println("Error parsing JSON [%v]", err)
		}
	}

	if tradeStruct.Party == tranOriginatorParty {
		notStruct.ParticipantID = tradeStruct.Counterparty

	} else if tradeStruct.Counterparty == tranOriginatorParty {
		notStruct.ParticipantID = tradeStruct.Party

	}

	//Notification for Party
	err = t.newNotificationEntry(stub, notStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Cash Adjustment Notification Entry failed", err)
		return err
	}

	var mpStruct MultiParty
	mpStruct.EventType = tranType
	mpStruct.MinCount = 1
	mpStruct.PendingCount = 1
	mpStruct.TransactionRef = sysReference
	mpStruct.ActionByUser = "RepoChaincode"

	err = t.newMultiParty(stub, mpStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Cash Adjustment MultiParty Entry failed", err)
		return err
	}

	return nil
}

func (t *multipartyHandler) newCollateralSubNotification(stub shim.ChaincodeStubInterface, sysReference string, repoJSON string, tranType string, mpChaincodeID string, repoChannelID string) error {

	fmt.Println("###### CollateralSubChaincode: function: newCollateralSubNotification ")

	var err error
	var arbitrary_json map[string]interface{}
	err = json.Unmarshal([]byte(repoJSON), &arbitrary_json)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}

	fmt.Println("Trade Data: %v", arbitrary_json["Trade"])
	jsonByte, err := json.Marshal(arbitrary_json["Trade"])
	jsonStr := convertArray(jsonByte)
	fmt.Println("Trade Data: %s", string(jsonStr))

	var tradeStruct TradeDetails
	err = json.Unmarshal([]byte(jsonStr), &tradeStruct)
	if err != nil {
		fmt.Println("Error parsing JSON [%v]", err)
	}

	var notStruct NotificationsDetails
	notStruct.NotificationType = tranType
	notStruct.NotificationStatus = "Pending"
	notStruct.NotificationData = repoJSON
	notStruct.NotificationDesc = "Collateral Substitution Approval"
	notStruct.TransactionRef = sysReference

	if tradeStruct.TranOriginatorParty == tradeStruct.Party {
		notStruct.ParticipantID = tradeStruct.Counterparty

	} else if tradeStruct.TranOriginatorParty == tradeStruct.Counterparty {
		notStruct.ParticipantID = tradeStruct.Party
	}

	//Notification for Party
	err = t.newNotificationEntry(stub, notStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("Collateral Sub New Notification Entry failed", err)
		return err
	}

	var mpStruct MultiParty
	mpStruct.EventType = tranType
	mpStruct.MinCount = 1
	mpStruct.PendingCount = 1
	mpStruct.TransactionRef = sysReference
	mpStruct.ActionByUser = "CollaterlSubChaincode"

	err = t.newMultiParty(stub, mpStruct, mpChaincodeID, repoChannelID)
	if err != nil {
		fmt.Println("CollateralSub MultiParty Entry failed", err)
		return err
	}

	return nil
}
