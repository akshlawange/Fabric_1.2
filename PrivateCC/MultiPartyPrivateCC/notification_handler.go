package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type NotificationsDetails struct {
	ObjectType         string `json:"ObjectType,omitempty"`
	ParticipantID      string `json:"ParticipantID,omitempty"`
	TransactionRef     string `json:"TransactionRef,omitempty"`
	NotificationType   string `json:"NotificationType,omitempty"` //New, Amend, Cancel, CollateralSubstitution, InterestPayment, CashAdjustment, Close
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
	EventType      string `json:"EventType,omitempty"`   //Repo New, Repo Amend, Repo Cancel, Collateral Substitution, Interest Payment, RepoClose
	EventStatus    string `json:"EventStatus,omitempty"` //Pending, Approved, Rejected
	MinCount       int    `json:"MinCount,omitempty"`
	PendingCount   int    `json:"PendingCount,omitempty"`
	ApproveCount   int    `json:"ApproveCount,omitempty"`
	RejectCount    int    `json:"RejectCount,omitempty"`
	DateTime       string `json:"DateTime,omitempty"`
	ActionByUser   string `json:"ActionByUser,omitempty"`
	ActiveInd      string `json:"ActiveInd,omitempty"`
}

//CollateralHandler provides APIs used to perform operations on CC's KV store
type notificationHandler struct {
}

// NewCollateralHandler create a new reference to CertHandler
func NewNotificationHandler() *notificationHandler {
	return &notificationHandler{}
}

func (t *notificationHandler) newNotificationCapture(stub shim.ChaincodeStubInterface, collection string, notiStruct NotificationsDetails) ([]byte, error) {

	fmt.Println("###### MultiPartyChaincode: function: newNotificationCapture ")
	fmt.Println("Insert notification= ", notiStruct.NotificationID)

	notiStruct.ObjectType = "NotificationsDetails"
	notiStruct.DateTime = time.Now().UTC().String()
	notiStruct.ActiveInd = "A"
	notiID := strconv.Itoa(notiStruct.NotificationID)
	compositeKey, _ := stub.CreateCompositeKey(notiStruct.ObjectType, []string{notiStruct.ParticipantID, notiStruct.NotificationType, notiStruct.TransactionRef, notiID})
	notiJSONBytes, _ := json.Marshal(notiStruct)
	fmt.Println("New Notificaiton Composite key", compositeKey)
	stub.PutPrivateData(collection, compositeKey, notiJSONBytes)

	return nil, nil
}

func (t *notificationHandler) updateNotification(stub shim.ChaincodeStubInterface, collection string, notiStruct NotificationsDetails) ([]byte, error) {

	fmt.Println("###### MultiPartyChaincode: function: updateNotification ")

	fmt.Println("Update notification= ", notiStruct.NotificationID)

	notiStruct.ObjectType = "NotificationsDetails"
	notiStruct.DateTime = time.Now().UTC().String()
	notiStruct.ActiveInd = "A"
	notiID := strconv.Itoa(notiStruct.NotificationID)
	compositeKey, _ := stub.CreateCompositeKey(notiStruct.ObjectType, []string{notiStruct.ParticipantID, notiStruct.NotificationType, notiStruct.TransactionRef, notiID})
	notiJSONBytes, _ := json.Marshal(notiStruct)

	stub.PutPrivateData(collection, compositeKey, notiJSONBytes)

	return nil, nil
}

func (t *notificationHandler) deactivateNotification(stub shim.ChaincodeStubInterface, collection string, notiStruct NotificationsDetails) ([]byte, error) {

	fmt.Println("###### MultiPartyChaincode: function: deactivateNotification ")

	fmt.Println("Deactive notification= ", notiStruct.NotificationID)

	notiStruct.ObjectType = "NotificationsDetails"
	notiStruct.DateTime = time.Now().UTC().String()
	//notiStruct.ActiveInd = "N"
	notiID := strconv.Itoa(notiStruct.NotificationID)
	compositeKey, _ := stub.CreateCompositeKey(notiStruct.ObjectType, []string{notiStruct.ParticipantID, notiStruct.NotificationType, notiStruct.TransactionRef, notiID})
	notiJSONBytes, _ := json.Marshal(notiStruct)

	stub.PutPrivateData(collection, compositeKey, notiJSONBytes)

	return nil, nil
}

func (t *notificationHandler) deleteNotification(stub shim.ChaincodeStubInterface, collection string, notiStruct NotificationsDetails) ([]byte, error) {

	fmt.Println("###### MultiPartyChaincode: function: deleteNotification ")

	fmt.Println("Delete notification=", notiStruct.NotificationID)

	notiStruct.ObjectType = "NotificationsDetails"
	notiStruct.DateTime = time.Now().UTC().String()
	notiStruct.ActiveInd = "A"
	notiID := strconv.Itoa(notiStruct.NotificationID)
	compositeKey, _ := stub.CreateCompositeKey(notiStruct.ObjectType, []string{notiStruct.ParticipantID, notiStruct.NotificationType, notiStruct.TransactionRef, notiID})

	stub.DelPrivateData(collection, compositeKey)

	return nil, nil
}

// queryNotification returns the record row matching a corresponding notification on the chaincode state table
func (t *notificationHandler) queryNotification(stub shim.ChaincodeStubInterface, collection string, participantID string, notificationType string, transactionRef string, notificationID string) ([]byte, error) {

	fmt.Println("###### MultiPartyChaincode: function: queryNotification ")

	if notificationID != "" && participantID != "" && notificationType != "" && transactionRef != "" {

		objectType := "NotificationsDetails"
		compositeKey, _ := stub.CreateCompositeKey(objectType, []string{participantID, notificationType, transactionRef, notificationID})		
		notiJSONBytes, _ := stub.GetPrivateData(collection, compositeKey)				
		return notiJSONBytes, nil
	}
	return nil, nil
}

// queryAllNotificationsByParticipant returns the record row matching a correponding Participant ID on the chaincode state table
func (t *notificationHandler) queryAllNotificationsByParticipant(stub shim.ChaincodeStubInterface, collection string, participantID string) ([]byte, error) {
	fmt.Println("###### MultiPartyChaincode: function: queryAllNotificationsByParticipant ")

	if participantID != "" {

		var attributes []string
		attributes = append(attributes, participantID)

		finaldata, err := utilHandler.readMultiJSON(stub, collection, "NotificationsDetails", attributes)
		if err != nil {
			fmt.Println("Error querying multi notifications", err)
			return nil, err
		}

		fmt.Println("Notification ::", string(finaldata))
		return finaldata, nil
	}
	return nil, nil
}

// queryAllNotificationsByParticipant returns the record row matching a correponding Participant ID on the chaincode state table
func (t *notificationHandler) queryAllNotifications(stub shim.ChaincodeStubInterface, collection string, participantID string, notificationType string) ([]byte, error) {
	fmt.Println("###### MultiPartyChaincode: function: queryAllNotificationsByParticipant ")

	if participantID != "" || notificationType != "" {

		var attributes []string
		attributes = append(attributes, participantID)
		attributes = append(attributes, notificationType)

		finaldata, err := utilHandler.readMultiJSON(stub, collection, "NotificationsDetails", attributes)
		if err != nil {
			fmt.Println("Error querying multi notifications", err)
			return nil, err
		}

		fmt.Println("Notification ::", string(finaldata))
		return finaldata, nil
	}
	return nil, nil
}

func (t *notificationHandler) newMultiPartyEvent(stub shim.ChaincodeStubInterface, collection string, mpStruct MultiParty) error {

	fmt.Println("###### MultiPartyChaincode: function: newMultiPartyEvent ")
	mpStruct.ObjectType = "MultiParty"
	mpStruct.EventStatus = "Pending"
	mpStruct.ApproveCount = 0
	mpStruct.RejectCount = 0
	mpStruct.DateTime = time.Now().UTC().String()
	mpStruct.ActiveInd = "A"

	compositeKey, _ := stub.CreateCompositeKey(mpStruct.ObjectType, []string{mpStruct.TransactionRef, mpStruct.EventType})
	mpJSONBytes, _ := json.Marshal(mpStruct)
	fmt.Println("New MultiPartyEvent Key:", compositeKey)
	stub.PutPrivateData(collection, compositeKey, mpJSONBytes)

	return nil
}

func (t *notificationHandler) delMultiPartyEvent(stub shim.ChaincodeStubInterface, collection string, mpStruct MultiParty) error {

	fmt.Println("###### MultiPartyChaincode: function: invoke ")
	fmt.Println("Delete Multi Party counter= ", mpStruct.TransactionRef)

	mpStruct.ObjectType = "MultiParty"
	compositeKey, _ := stub.CreateCompositeKey(mpStruct.ObjectType, []string{mpStruct.TransactionRef, mpStruct.EventType})

	stub.DelPrivateData(collection, compositeKey)

	return nil
}

func (t *notificationHandler) updateMultiPartyEvent(stub shim.ChaincodeStubInterface, collection string, mpStruct MultiParty) error {

	fmt.Println("###### MultiPartyChaincode: function: invoke ")
	fmt.Println("update Multi Party counter= ", mpStruct.TransactionRef)

	mpStruct.ObjectType = "MultiParty"
	mpStruct.DateTime = time.Now().UTC().String()
	mpStruct.ActiveInd = "A"
	compositeKey, _ := stub.CreateCompositeKey(mpStruct.ObjectType, []string{mpStruct.TransactionRef, mpStruct.EventType})
	mpJSONBytes, _ := json.Marshal(mpStruct)

	stub.PutPrivateData(collection, compositeKey, mpJSONBytes)

	return nil
}

func (t *notificationHandler) queryMultiPartyEvent(stub shim.ChaincodeStubInterface, collection string, transactionRef string, eventType string) ([]byte, error) {

	fmt.Println("###### MultiPartyChaincode: function: invoke ")

	if transactionRef != "" || eventType != "" {
		var attributes []string
		attributes = append(attributes, transactionRef)
		attributes = append(attributes, eventType)

		mpJSONBytes, err := utilHandler.readSingleJSON(stub, collection, "MultiParty", attributes)
		if err != nil {
			fmt.Println("Error querying multi party count", err)
			return nil, err
		}
		fmt.Println("Multiparty count :: ", string(mpJSONBytes))

		return mpJSONBytes, nil
	}
	return nil, nil
}
