package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type MultiPartyChaincode struct {
}

var notiHandler = NewNotificationHandler()
var aHandler = NewAssetHandler()
var utilHandler = NewUtilityHandler()

type NotificationIDToken struct {
	NotificationIDCount int32 `json:"NotificationIDCount,string,omitempty"`
}

func (t *MultiPartyChaincode) newNotificationCapture(stub shim.ChaincodeStubInterface, collection string, notification string) pb.Response {

	var err error
	fmt.Println("###### MultiPartyChaincode: function: newNotificationCapture ")

	var notStruct NotificationsDetails
	err = json.Unmarshal([]byte(notification), &notStruct)
	if err != nil {
		return shim.Error("Error in unmarshalling notification")
	}

	fmt.Println("Generate Notifcation ID")
	id, err := t.generateNewNotificationID(stub)
	if err != nil {
		return shim.Error("Error in generating notification id")
	}

	notStruct.NotificationID = int(id)
	notStruct.NotificationStatus = "Pending"
	notStruct.Comment = "Pending Notification"

	notiHandler.newNotificationCapture(stub, collection, notStruct)

	return shim.Success([]byte("Notification capture is successfully!"))
}

func (t *MultiPartyChaincode) approveNotification(stub shim.ChaincodeStubInterface, collection string, participantID string, notificationType string, transactionRef string, notificationID string, userID string, assetOwnershipChaincodeID string, channelID string) pb.Response {

	var err error	
	fmt.Println("###### MultiPartyChaincode: function: approveNotification ")

	var notStruct NotificationsDetails
	notiJSON, err := notiHandler.queryNotification(stub, collection, participantID, notificationType, transactionRef, notificationID)
	if err != nil {
		fmt.Println("Error in querying notification: ", err)
		return shim.Error("Error in querying notification")
	}

	if string(notiJSON) != "" {
		err = json.Unmarshal([]byte(notiJSON), &notStruct)
		if err != nil {
			fmt.Println("Error in unmarshalling notification: ", err)
			return shim.Error("Error in unmarshalling notification")
		}
		//Approve Notification
		notStruct.NotificationStatus = "Approved"
		notStruct.Comment = "Approved by User"
		notStruct.ActiveInd = "A"
		notStruct.ActionByUser = userID

		_, err = notiHandler.updateNotification(stub, collection, notStruct)
		if err != nil {
			fmt.Println("Error in updating notification: ", err)
			return shim.Error("Error in updating notification")
		}

		eventStatus, err := t.updateMultiParty(stub, collection, transactionRef, notificationType, "APPROVE")
		fmt.Println("Event Status : ", eventStatus)
		if err != nil {
			fmt.Println("Error in updating multiparty count: ", err)
			return shim.Error("Error in updating multiparty count")
		}

		if eventStatus == "Approved" && notificationType == "TokenIssuance"  {
			fmt.Println("Deploying an Asset : %s", notStruct.NotificationData)

			err = aHandler.deployAssetOwnership(stub, notStruct.NotificationData, notStruct.NotificationType, assetOwnershipChaincodeID, channelID)
			if err != nil {
				fmt.Println("Repo deal Deployment Failed:", err)
				return shim.Error("Repo deal Deployment Failed")
			}
		
		} else if eventStatus == "Approved" && notificationType == "TokenRedemption" {

			fmt.Println("Redeeming an asset : %s", notStruct.NotificationData)
			err = aHandler.deployAssetOwnership(stub, notStruct.NotificationData, notStruct.NotificationType, assetOwnershipChaincodeID, channelID)
			if err != nil {
				fmt.Println("Repo deal cancel Failed:", err)
				return shim.Error("Repo deal cancel Failed")
			}
		}

		return shim.Success([]byte("Notification Approval is successfully!"))
	}
	return shim.Success([]byte("Notification is NOT found!!"))
}

func (t *MultiPartyChaincode) rejectNotification(stub shim.ChaincodeStubInterface, collection string, participantID string, notificationType string, transactionRef string, notificationID string, userID string, assetOwnershipChaincodeID string, channelID string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: rejectNotification ")
	var err error
	//Parameters: ParticipantID, NotificationType, TransactionRef, NotificationID

	var notStruct NotificationsDetails
	notiJSON, err := notiHandler.queryNotification(stub, collection, participantID, notificationType, transactionRef, notificationID)
	if err != nil {
		fmt.Println("Error in querying notification: ", err)
		return shim.Error("Error in querying notification")
	}
	if string(notiJSON) != "" {
		err = json.Unmarshal([]byte(notiJSON), &notStruct)
		if err != nil {
			fmt.Println("Error in unmarshalling date: ", err)
			return shim.Error("Error in unmarshalling notification")
		}
		//Approve Notification
		notStruct.NotificationStatus = "Rejected"
		notStruct.Comment = "Rejected by User"
		notStruct.ActiveInd = "A"
		notStruct.ActionByUser = userID

		_, err = notiHandler.updateNotification(stub, collection, notStruct)
		if err != nil {
			fmt.Println("Error in updating notification: ", err)
			return shim.Error("Error in updating notification")
		}

		eventStatus, err := t.updateMultiParty(stub, collection, transactionRef, notificationType, "REJECT")
		fmt.Println("Event Status : ", eventStatus)
		if err != nil {
			fmt.Println("Error in updating multiparty count: ", err)
			return shim.Error("Error in updating multiparty count")
		}		
		
		return shim.Success([]byte("Notification is rejected successfully"))
	}
	return shim.Success([]byte("Notification is NOT found!!"))
}

func (t *MultiPartyChaincode) newMultiParty(stub shim.ChaincodeStubInterface, collection string, mpEvent string) pb.Response {
	fmt.Println("###### MultiPartyChaincode: function: newMultiParty ")
	var err error
	var mpStruct MultiParty
	err = json.Unmarshal([]byte(mpEvent), &mpStruct)
	if err != nil {
		fmt.Println("Error unmarshalling new repo event", err)
	}

	err = notiHandler.newMultiPartyEvent(stub, collection, mpStruct)
	if err != nil {
		fmt.Println("Error in creating multi party event: %v", err)
		return shim.Error("Error in creating multi party event")
	}

	return shim.Success([]byte("MultiParty event is captured successfully"))
}

func (t *MultiPartyChaincode) updateMultiParty(stub shim.ChaincodeStubInterface, collection string, transactionRef string, eventType string, action string) (string, error) {
	fmt.Println("###### MultiPartyChaincode: function: updateMultiParty ")

	var mpStruct MultiParty
	//query--->
	multiJSON, err := notiHandler.queryMultiPartyEvent(stub, collection, transactionRef, eventType)
	if err != nil {
		fmt.Println("Error in querying multi party count: %v", err)
		return "Pending", err
	}

	if string(multiJSON) != "" {
		err = json.Unmarshal([]byte(multiJSON), &mpStruct)
		if err != nil {
			fmt.Println("Error in unmarshalling multi party: %v", err)
			return "Pending", err
		}
	} else {
		return "Error", nil
	}

	if mpStruct.PendingCount > 0 {
		mpStruct.PendingCount = mpStruct.PendingCount - 1

		if action == "APPROVE" {
			mpStruct.ApproveCount = mpStruct.ApproveCount + 1
		} else if action == "REJECT" {
			mpStruct.RejectCount = mpStruct.RejectCount + 1
		}

		if mpStruct.ApproveCount >= mpStruct.MinCount {
			mpStruct.EventStatus = "Approved"
			mpStruct.ActiveInd = "N"
			err = notiHandler.updateMultiPartyEvent(stub, collection, mpStruct)
			if err != nil {
				fmt.Println("Error in deleteing multi party count: %v", err)
				return "Pending", err
			}
			return "Approved", nil
			//WHAT ABOUT OTHER PENDING NOTIFICATION??
		} else if mpStruct.RejectCount > (mpStruct.PendingCount - mpStruct.MinCount) {
			mpStruct.EventStatus = "Rejected"
			mpStruct.ActiveInd = "N"
			err = notiHandler.updateMultiPartyEvent(stub, collection, mpStruct)
			if err != nil {
				fmt.Println("Error in deleteing multi party count: %v", err)
				return "Pending", err
			}
			//WHAT ABOUT OTHER PENDING NOTIFICATION??
			return "Rejected", nil
		}
	}
	return "Pending", nil
}

func (t *MultiPartyChaincode) queryMultiParty(stub shim.ChaincodeStubInterface, collection string,  transactionRef string, eventType string) (int, error) {
	fmt.Println("###### MultiPartyChaincode: function: queryMultiParty ")

	countJSON, err := notiHandler.queryMultiPartyEvent(stub, collection, transactionRef, eventType)
	if err != nil {
		fmt.Println("Error in querying multi party event: %v", err)
		return -1, err
	}

	var mpStruct MultiParty
	err = json.Unmarshal([]byte(countJSON), &mpStruct)
	if err != nil {
		fmt.Println("Error in unmarshalling multi party:", err)
		return -1, err

	}
	return mpStruct.PendingCount, nil
}

func (t *MultiPartyChaincode) queryAllNotification(stub shim.ChaincodeStubInterface, collection string, participantID string, notificationType string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryAllNotification ")
	fmt.Println("Querying for : [%s][%s]", string(participantID), string(notificationType))

	notVal, err := notiHandler.queryAllNotifications(stub, collection, participantID, notificationType)
	if err != nil {
		fmt.Println("Failed to query a notifications ", err)
		shim.Error("Failed to query a notifications")
	}
	fmt.Println("Query sec return %s", notVal)

	return shim.Success([]byte(notVal))
}

func (t *MultiPartyChaincode) queryNotification(stub shim.ChaincodeStubInterface, collection string, participantID string, notificationType string, transactionRef string, notificationID string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryNotification ")
	//Parameters: ParticipantID, NotificationType, TransactionRef, NotificationID

	notVal, err := notiHandler.queryNotification(stub, collection, participantID, notificationType, transactionRef, notificationID)
	if err != nil {
		fmt.Println("Failed to query a notification row ", err)
		return shim.Error("Failed to query a notification row")
	}
	fmt.Println("Query sec return %s", string(notVal))

	return shim.Success([]byte(notVal))
}

/*
func (t *MultiPartyChaincode) queryNotificationsByFilter(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryNotificationsByFilter ")

	var participantID string
	var notificationType string
	var notificationStatus string

	participantID = args[1]
	notificationType = args[2]
	notificationStatus = args[3]

	queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"NotificationsDetails\", \"ParticipantID\":\"%s\", \"NotificationType\":\"%s\", \"NotificationStatus\":\"%s\"}}", participantID, notificationType, notificationStatus)
	//queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"NotificationDetails\"}}")

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
} */

func (t *MultiPartyChaincode) queryNotificationsByFilter(stub shim.ChaincodeStubInterface, collection string, queryString string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryNotificationsByFilter ", queryString)

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, collection, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func (t *MultiPartyChaincode) queryNotificationsByTransactionRef(stub shim.ChaincodeStubInterface, collection string, transactionRef string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryNotificationsByTransactionRef ")

	queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"NotificationsDetails\", \"TransactionRef\":\"%s\", \"NotificationStatus\":\"Pending\"}}", transactionRef)

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, collection, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func (t *MultiPartyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: Invoke ")

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Function: %v %v", function, args)

	if len(args) < 2 {
		return shim.Error("Not enough args passed")
	}

	if function[0:1] == "i" {
		return t.invoke(stub, args[0], args) // old invoke function
	}
	if function[0:1] == "q" {
		return t.query(stub, args[0], args) // old query function
	}
	return shim.Error("Invoke: Invalid Function Name - function names begin with a q or i")
}

func (t *MultiPartyChaincode) invoke(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### MultiPartyChaincode: function: invoke ")
	fmt.Println("length JSON Data: %v %v", args[0], len(args))

	if function == "newNotificationCapture" {
		collection := args[1]
		notification := args[2]
		return t.newNotificationCapture(stub, collection, notification)

	} else if function == "approveNotification" {
		collection := args[1]
		participantID := args[2]
		notificationType := args[3]
		transactionRef := args[4]
		notificationID := args[5]
		userID := args[6]
		assetOwnershipChaincodeID := args[7]
		channelID := args[8]
		return t.approveNotification(stub, collection, participantID, notificationType, transactionRef, notificationID, userID, assetOwnershipChaincodeID, channelID)

	} else if function == "rejectNotification" {
		collection := args[1]
		participantID := args[2]
		notificationType := args[3]
		transactionRef := args[4]
		notificationID := args[5]
		userID := args[6]
		assetOwnershipChaincodeID := args[7]
		channelID := args[8]
		return t.rejectNotification(stub, collection, participantID, notificationType, transactionRef, notificationID, userID, assetOwnershipChaincodeID, channelID)

	} else if function == "newMultiParty" {
		collection := args[1]
		mpevent := args[2]
		return t.newMultiParty(stub, collection, mpevent)

	}

	return shim.Error("Received unknown function query invocation with function")
}

func (t *MultiPartyChaincode) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### MultiPartyChaincode: function: query ")

	if function == "queryAllNotification" {
		collection := args[1]
		participantID := args[2]
		notificationType := args[3]
		return t.queryAllNotification(stub, collection, participantID, notificationType)

	} else if function == "queryNotification" {
		collection := args[1]
		participantID := args[2]
		notificationType := args[3]
		transactionRef := args[4]
		notificationID := args[5]
		return t.queryNotification(stub, collection, participantID, notificationType, transactionRef, notificationID)

	} else if function == "queryNotificationsByFilter" {
		collection := args[1]
		queryString := args[2]
		return t.queryNotificationsByFilter(stub, collection, queryString)

	} else if function == "queryNotificationsByTransactionRef" {
		collection := args[1]
		transactionRef := args[2]
		return t.queryNotificationsByTransactionRef(stub, collection, transactionRef)
	
	}

	return shim.Error("Received unknown function query invocation with function")
}

func (t *MultiPartyChaincode) generateNewNotificationID(stub shim.ChaincodeStubInterface) (int32, error) {

	fmt.Println("###### MultiPartyChaincode: function: generateNewNotificationID ")
	objectType := "NotificationToken"

	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(objectType, []string{"NotificationIDCount"})
	extJSONBytes, _ := stub.GetState(compositeKey)

	var notificationTokenStruct NotificationIDToken
	err := json.Unmarshal([]byte(extJSONBytes), &notificationTokenStruct)
	if err != nil {
		fmt.Println("Error parsing JSON [%v]", err)
	}

	newToken := notificationTokenStruct.NotificationIDCount
	notificationTokenStruct.NotificationIDCount = notificationTokenStruct.NotificationIDCount + 1
	extJSONBytes, _ = json.Marshal(notificationTokenStruct)
	err = stub.PutState(compositeKey, extJSONBytes)
	if err != nil {
		return 0, errors.New("Error in updating ref token state")
	}

	return newToken, nil
}

func (t *MultiPartyChaincode) initialiseNewNotificationID(stub shim.ChaincodeStubInterface) error {

	fmt.Println("###### MultiPartyChaincode: function: initialiseNewNotificationID ")

	objectType := "NotificationToken"

	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(objectType, []string{"NotificationIDCount"})

	var notificationTokenStruct NotificationIDToken
	notificationTokenStruct.NotificationIDCount = 100
	extJSONBytes, _ := json.Marshal(notificationTokenStruct)
	err := stub.PutState(compositeKey, extJSONBytes)
	if err != nil {
		return errors.New("Error in updating notification token state")
	}

	return nil
}

func (t *MultiPartyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	var err error
	fmt.Println("###### MultiPartyChaincode: function: Init ")

	err = t.initialiseNewNotificationID(stub)
	if err != nil {
		return shim.Error("Error Initialising Token")
	}

	return shim.Success(nil)
}

func main() {
	//	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(MultiPartyChaincode))
	if err != nil {
		fmt.Println("Error starting MultiPartyChaincode: %s", err)
	}

}
