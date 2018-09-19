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
var rHandler = NewRepoHandler()
var utilHandler = NewUtilityHandler()
var repoChainID = "RepoDealCC"
var repoChannelID = "repochannel"

type NotificationIDToken struct {
	NotificationIDCount int32 `json:"NotificationIDCount,string,omitempty"`
}

type TradeDetails struct {
	ObjectType                string `json:"ObjectType,omitempty"`
	ProcessingSystemReference string `json:"ProcessingSystemReference,omitempty"`
	FOReference               string `json:"FOReference,omitempty"`
	ExternalReference         string `json:"ExternalReference,omitempty"`
	Party                     string `json:"Party,omitempty"`                 // PARTICIPANT_ID ON THE PARTY JSON
	PartyCustodian            string `json:"PartyCustodian,omitempty"`        // PARTICIPANT_ID ON THE PARTY CUSTODIAN JSON
	Counterparty              string `json:"Counterparty,omitempty"`          // PARTICIPANT_ID ON THE COUNTERPARTY JSON
	CounterpartyCustodian     string `json:"CounterpartyCustodian,omitempty"` // PARTICIPANT_ID ON THE COUNTERPARTY CUSTODIAN JSONs
	TradeType                 string `json:"TradeType,omitempty"`             // BORROW VS LOAN, REPO etc.
	TransactionType           string `json:"TransactionType,omitempty"`       // FICC, NON-FICC, INTERNAL
	TransactionStatus         string `json:"TransactionStatus,omitempty"`     // NEW, AMEND, CANCEL
	RepoStatus                string `json:"RepoStatus,omitempty"`            // PENDAPPROVAL, APPROVED, --REPO DEAL CC STATUS: OPENLEGPEND, OPENLEGSETTLED, CLOSELEGPEND, CLOSELEGSETTLED
	TransactionDate           string `json:"TransactionDate,omitempty"`
	TransactionTimestamp      string `json:"TransactionTimestamp,omitempty"`
	EffectiveDate             string `json:"EffectiveDate,omitempty"`
	ContractualValueDate      string `json:"ContractualValueDate,omitempty"`
	CloseEventDate            string `json:"CloseEventDate,omitempty"`
	PlaceOfTrade              string `json:"PlaceOfTrade,omitempty"`
	BaseCurrency              string `json:"BaseCurrency,omitempty"`
	SettleCurrency            string `json:"SettleCurrency,omitempty"`
	TotalCashAmount           string `json:"TotalCashAmount,omitempty"` // CASH AMOUNT = SUM OF ALL COLLATERAL CASH AMT
	CurrentFinancingRate      string `json:"CurrentFinancingRate,omitempty"`
	DayCount                  string `json:"DayCount,omitempty"`
	InterestDays              int64  `json:"InterestDays,string,omitempty"` // CASH Interest Acc days
	AccruedInterest           string `json:"AccruedInterest,omitempty"`
	TotalPaidInterest         string `json:"TotalPaidInterest,omitempty"`
	PlaceOfSettlement         string `json:"PlaceOfSettlement,omitempty"`
	SettlementTerms           string `json:"SettlementTerms,omitempty"`
	TransactionState          string `json:"TransactionState,omitempty"`
	Comment                   string `json:"Comment,omitempty"`
	LastUpdatedUser           string `json:"LastUpdatedUser,omitempty"`
	DateTime                  string `json:"DateTime,omitempty"`
	Version                   int    `json:"Version,string,omitempty"`
	ActiveInd                 string `json:"ActiveInd,omitempty"`
}

func (t *MultiPartyChaincode) newNotificationCapture(stub shim.ChaincodeStubInterface, notification string) pb.Response {

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

	notiHandler.newNotificationCapture(stub, notStruct)

	return shim.Success([]byte("Notification capture is successfully!"))
}

/*
func (t *MultiPartyChaincode) deactivateNotification(stub shim.ChaincodeStubInterface, notification string) error {

	fmt.Println("###### MultiPartyChaincode: function: deactivateNotification ")

	var notStruct NotificationsDetails

	notiJSON, err := t.queryNotification(stub, notification)
	if err != nil {
		fmt.Println("Error in querying notification: %v", err)
		return err
	}
	err = json.Unmarshal([]byte(notiJSON), &notStruct)
	if err != nil {
		fmt.Println("Error unmarshalling repo notification", err)
	}

	notStruct.Active = "N"

	fmt.Println("Participant ID : %v", notStruct.ParticipantID)
	fmt.Println("Notification Type : %v", notStruct.NotificationType)
	fmt.Println("Notification Desc : %v", notStruct.NotificationDesc)
	fmt.Println("Notification Data : %v", notStruct.NotificationData)
	fmt.Println("Notification Active : %v", notStruct.Active)

	_, err = notiHandler.deactivateNotification(stub, notStruct)
	if err != nil {
		fmt.Println("Error in deactivating notification: %v", err)
		return err
	}

	return nil
}*/

func (t *MultiPartyChaincode) approveNotification(stub shim.ChaincodeStubInterface, participantID string, notificationType string, transactionRef string, notificationID string, userID string, repoChaincodeID string, collSubChaincodeID string, repoChannelID string) pb.Response {

	var err error
	var repoStatus string
	fmt.Println("###### MultiPartyChaincode: function: approveNotification ")

	//CHECK IF MULTIPARTY EVENT IS NOT EXPIRED?

	//CHECK IF REPO DEAL (Repo Status) is NOT in OPENLEGPEND or CLOSELEGPEND or COLLSUBPEND or INTPAYMENTPEND
	if notificationType != "New" {
		repoStatusByte, err := rHandler.getRepoDealStatus(stub, transactionRef, repoChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error reading Repo Deal Status", err)
		}
		repoStatus = string(repoStatusByte)

		if string(repoStatus) == "OPENLEGPEND" || string(repoStatus) == "CLOSELEGPEND" || string(repoStatus) == "COLLSUBPEND" || string(repoStatus) == "INTPAYMENTPEND" {
			return shim.Success([]byte("Repo Deal is not in expected status, pls retry later"))
		}

	}
	var notStruct NotificationsDetails
	notiJSON, err := notiHandler.queryNotification(stub, participantID, notificationType, transactionRef, notificationID)
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

		_, err = notiHandler.updateNotification(stub, notStruct)
		if err != nil {
			fmt.Println("Error in updating notification: ", err)
			return shim.Error("Error in updating notification")
		}

		eventStatus, err := t.updateMultiParty(stub, transactionRef, notificationType, "APPROVE")
		fmt.Println("Event Status : ", eventStatus)
		if err != nil {
			fmt.Println("Error in updating multiparty count: ", err)
			return shim.Error("Error in updating multiparty count")
		}

		if eventStatus == "Approved" && notificationType == "New" || notificationType == "Amend" {
			fmt.Println("Deploying RepoDeal : %s", notStruct.NotificationData)
			err = rHandler.deployRepoDeal(stub, transactionRef, notStruct.NotificationData, notStruct.NotificationType, "Approved", repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Repo deal Deployment Failed:", err)
				return shim.Error("Repo deal Deployment Failed")
			}
			notificationString := "{\"transactionRef\":\"" + transactionRef + "\",\"repoJSON\":" + notStruct.NotificationData + "}"

			if repoStatus != "OPENLEGSETTLED" || repoStatus != "REJECTED" {
				f := "invoke"
				invokeArgs := util.ToChaincodeArgs(f, "initiateRepoDealSettlement", string(collection), notificationString)
				response := stub.InvokeChaincode(repoChainID, invokeArgs, repoChannelID)
				if response.Status != shim.OK {
					errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
					fmt.Printf(errStr)
					return errors.New(errStr)
				}
				// fmt.Println("Set Event InitiateOpenDealSettlement: %s", notificationString)
				// err = stub.SetEvent("InitiateOpenDealSettlement", []byte(notificationString))
				// if err != nil {
				// 	fmt.Println("OpenDealSettlement SetEvent Error", err)
				// }
			}

		} else if eventStatus == "Approved" && notificationType == "Cancel" {
			fmt.Println("Cancel RepoDeal : %s", transactionRef)
			err = rHandler.deployRepoDeal(stub, transactionRef, notStruct.NotificationData, notStruct.NotificationType, "Approved", repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Repo deal cancel Failed:", err)
				return shim.Error("Repo deal cancel Failed")
			}

			eventData, err := rHandler.getRepoDeal(stub, transactionRef, repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Error querying repo deal", err)
			}

			err = stub.SetEvent("InitiateCancelDealSettlement", []byte(eventData))
			if err != nil {
				fmt.Println("CancelDealSettlement SetEvent Error", err)
			}

		} else if eventStatus == "Approved" && notificationType == "Close" {
			fmt.Println("Closing RepoDeal : %s", transactionRef)
			err = rHandler.deployRepoDeal(stub, transactionRef, notStruct.NotificationData, notStruct.NotificationType, "Approved", repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Repo deal close Failed:", err)
				return shim.Error("Repo deal close Failed")
			}

			eventData, err := rHandler.getRepoDeal(stub, transactionRef, repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Error querying repo deal", err)
			}

			err = stub.SetEvent("InitiateCloseDealSettlement", []byte(eventData))
			if err != nil {
				fmt.Println("CloseDealSettlement SetEvent Error", err)
			}

		} else if eventStatus == "Approved" && notificationType == "CollateralSubstitution" {
			fmt.Println("Collateral Substitution RepoDeal : %s", notStruct.NotificationData)
			/*	err = rHandler.deployCollateralSub(stub, transactionRef, notStruct.NotificationData, repoChaincodeID, collSubChaincodeID, repoChannelID)
				if err != nil {
					fmt.Println("Repo deal Substitution Failed:", err)
					return shim.Error("Repo deal Substitution Failed")
				}*/
			err = rHandler.deployRepoDeal(stub, transactionRef, notStruct.NotificationData, notStruct.NotificationType, "Approved", repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Repo deal Collateral Sub Failed:", err)
				return shim.Error("Repo deal Collateral Sub Failed")
			}

			err = stub.SetEvent("InitiateCollaterlSubSettlement", []byte(notStruct.NotificationData))
			if err != nil {
				fmt.Println("CloseDealSettlement SetEvent Error", err)
			}

		} else if eventStatus == "Approved" && notificationType == "InterestPayment" {
			fmt.Println("InterestPayment : %s", notStruct.NotificationData)
			/*
				err = rHandler.deployRepoInterestPayment(stub, transactionRef, notStruct.NotificationData, repoChaincodeID, repoChannelID)
				if err != nil {
					fmt.Println("Repo deal Interim Payment Failed:", err)
					return shim.Error("Repo deal Interim Payment Failed")
				} */

			err = rHandler.deployRepoDeal(stub, transactionRef, notStruct.NotificationData, notStruct.NotificationType, "Approved", repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Repo deal Collateral Sub Failed:", err)
				return shim.Error("Repo deal Collateral Sub Failed")
			}
			/*
				eventData, err := rHandler.getRepoDeal(stub, transactionRef, repoChaincodeID, repoChannelID)
				if err != nil {
					fmt.Println("Error querying repo deal", err)
				}

				eventData = []byte("{" + string(eventData) + "},{\"PaymentAmount\":\"" + notStruct.NotificationData + "\"}")
				fmt.Println("Payment EventData: ", eventData)
			*/
			err = stub.SetEvent("InitiateInterestPayment", []byte(notStruct.NotificationData))
			if err != nil {
				fmt.Println("InterestPayment SetEvent Error", err)
			}
		} else if eventStatus == "Approved" && notificationType == "CashAdjustment" {
			fmt.Println("CashAdjustment : %s", notStruct.NotificationData)

			err = rHandler.deployRepoDeal(stub, transactionRef, notStruct.NotificationData, notStruct.NotificationType, "Approved", repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Repo deal Collateral Sub Failed:", err)
				return shim.Error("Repo deal Collateral Sub Failed")
			}
			err = stub.SetEvent("InitiateCashAdjustment", []byte(notStruct.NotificationData))
			if err != nil {
				fmt.Println("CashAdjustment SetEvent Error", err)
			}
		}

		return shim.Success([]byte("Notification Approval is successfully!"))
	}
	return shim.Success([]byte("Notification is NOT found!!"))
}

func (t *MultiPartyChaincode) rejectNotification(stub shim.ChaincodeStubInterface, participantID string, notificationType string, transactionRef string, notificationID string, userID string, repoChaincodeID string, collSubChaincodeID string, repoChannelID string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: rejectNotification ")
	var err error
	//Parameters: ParticipantID, NotificationType, TransactionRef, NotificationID

	//CHECK IF MULTIPARTY EVENT IS NOT EXPIRED?

	//CHECK IF REPO DEAL (Repo Status) is NOT in OPENLEGPEND or CLOSELEGPEND or COLLSUBPEND or INTPAYMENTPEND
	if notificationType != "New" {
		repoStatus, err := rHandler.getRepoDealStatus(stub, transactionRef, repoChaincodeID, repoChannelID)
		if err != nil {
			fmt.Println("Error reading Repo Deal Status", err)
		}

		if string(repoStatus) == "OPENLEGPEND" || string(repoStatus) == "CLOSELEGPEND" || string(repoStatus) == "COLLSUBPEND" || string(repoStatus) == "INTPAYMENTPEND" {
			return shim.Success([]byte("Repo Deal is not in expected status, pls retry later"))
		}

	}
	var notStruct NotificationsDetails
	notiJSON, err := notiHandler.queryNotification(stub, participantID, notificationType, transactionRef, notificationID)
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

		_, err = notiHandler.updateNotification(stub, notStruct)
		if err != nil {
			fmt.Println("Error in updating notification: ", err)
			return shim.Error("Error in updating notification")
		}

		eventStatus, err := t.updateMultiParty(stub, transactionRef, notificationType, "REJECT")
		fmt.Println("Event Status : ", eventStatus)
		if err != nil {
			fmt.Println("Error in updating multiparty count: ", err)
			return shim.Error("Error in updating multiparty count")
		}
		
		//notificationString := "{\"transactionRef\":\"" + transactionRef + "\",\"repoJSON\":" + notStruct.NotificationData + "}"

		if eventStatus == "Rejected" && notificationType == "New" {
			fmt.Println("Deploying RepoDeal : %s", notStruct.NotificationData)
			err = rHandler.deployRepoDeal(stub, transactionRef, notStruct.NotificationData, notStruct.NotificationType, "Rejected", repoChaincodeID, repoChannelID)
			if err != nil {
				fmt.Println("Repo deal Deployment Failed:", err)
				return shim.Error("Repo deal Deployment Failed")
			}

		} else if eventStatus == "Rejected" && notificationType == "Amend" || notificationType == "Cancel" {
			fmt.Println("Deploying RepoDeal : %s", notStruct.NotificationData)
			//DO NOTHING
			

		} else if eventStatus == "Rejected" && notificationType == "Close" {
			fmt.Println("Closing RepoDeal : %s", notStruct.NotificationData)
			//DO NOTHING


		} else if eventStatus == "Rejected" && notificationType == "CollateralSubstitution" {
			fmt.Println("Collateral Substitution RepoDeal : %s", notStruct.NotificationData)
			//DO NOTHING

		} else if eventStatus == "Rejected" && notificationType == "InterestPayment" {
			fmt.Println("InterestPayment : %s", notStruct.NotificationData)
			//DO NOTHING

		} else if eventStatus == "Rejected" && notificationType == "CashAdjustment" {
			fmt.Println("InterestPayment : %s", notStruct.NotificationData)
			//DO NOTHING
		}
		return shim.Success([]byte("Notification is rejected successfully"))
	}
	return shim.Success([]byte("Notification is NOT found!!"))
}

func (t *MultiPartyChaincode) newMultiParty(stub shim.ChaincodeStubInterface, mpEvent string) pb.Response {
	fmt.Println("###### MultiPartyChaincode: function: newMultiParty ")
	var err error
	var mpStruct MultiParty
	err = json.Unmarshal([]byte(mpEvent), &mpStruct)
	if err != nil {
		fmt.Println("Error unmarshalling new repo event", err)
	}

	err = notiHandler.newMultiPartyEvent(stub, mpStruct)
	if err != nil {
		fmt.Println("Error in creating multi party event: %v", err)
		return shim.Error("Error in creating multi party event")
	}

	return shim.Success([]byte("MultiParty event is captured successfully"))
}

func (t *MultiPartyChaincode) updateMultiParty(stub shim.ChaincodeStubInterface, transactionRef string, eventType string, action string) (string, error) {
	fmt.Println("###### MultiPartyChaincode: function: updateMultiParty ")

	var mpStruct MultiParty
	//query--->
	multiJSON, err := notiHandler.queryMultiPartyEvent(stub, transactionRef, eventType)
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
			err = notiHandler.updateMultiPartyEvent(stub, mpStruct)
			if err != nil {
				fmt.Println("Error in deleteing multi party count: %v", err)
				return "Pending", err
			}
			return "Approved", nil
			//WHAT ABOUT OTHER PENDING NOTIFICATION??
		} else if mpStruct.RejectCount > (mpStruct.PendingCount - mpStruct.MinCount) {
			mpStruct.EventStatus = "Rejected"
			mpStruct.ActiveInd = "N"
			err = notiHandler.updateMultiPartyEvent(stub, mpStruct)
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

func (t *MultiPartyChaincode) queryMultiParty(stub shim.ChaincodeStubInterface, transactionRef string, eventType string) (int, error) {
	fmt.Println("###### MultiPartyChaincode: function: queryMultiParty ")

	countJSON, err := notiHandler.queryMultiPartyEvent(stub, transactionRef, eventType)
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

func (t *MultiPartyChaincode) queryAllNotification(stub shim.ChaincodeStubInterface, participantID string, notificationType string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryAllNotification ")
	fmt.Println("Querying for : [%s][%s]", string(participantID), string(notificationType))

	notVal, err := notiHandler.queryAllNotifications(stub, participantID, notificationType)
	if err != nil {
		fmt.Println("Failed to query a notifications ", err)
		shim.Error("Failed to query a notifications")
	}
	fmt.Println("Query sec return %s", notVal)

	return shim.Success([]byte(notVal))
}

func (t *MultiPartyChaincode) queryNotification(stub shim.ChaincodeStubInterface, participantID string, notificationType string, transactionRef string, notificationID string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryNotification ")
	//Parameters: ParticipantID, NotificationType, TransactionRef, NotificationID

	notVal, err := notiHandler.queryNotification(stub, participantID, notificationType, transactionRef, notificationID)
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
func (t *MultiPartyChaincode) queryNotificationsByFilter(stub shim.ChaincodeStubInterface, queryString string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryNotificationsByFilter ", queryString)

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func (t *MultiPartyChaincode) queryNotificationsByTransactionRef(stub shim.ChaincodeStubInterface, transactionRef string) pb.Response {

	fmt.Println("###### MultiPartyChaincode: function: queryNotificationsByTransactionRef ")

	queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"NotificationsDetails\", \"TransactionRef\":\"%s\", \"NotificationStatus\":\"Pending\"}}", transactionRef)

	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
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
		notification := args[1]
		return t.newNotificationCapture(stub, notification)
	} else if function == "approveNotification" {
		participantID := args[1]
		notificationType := args[2]
		transactionRef := args[3]
		notificationID := args[4]
		userID := args[5]
		repoChaincodeID := args[6]
		collSubChaincodeID := args[7]
		repoChannelID := args[8]
		return t.approveNotification(stub, participantID, notificationType, transactionRef, notificationID, userID, repoChaincodeID, collSubChaincodeID, repoChannelID)
	} else if function == "rejectNotification" {
		participantID := args[1]
		notificationType := args[2]
		transactionRef := args[3]
		notificationID := args[4]
		userID := args[5]
		repoChaincodeID := args[6]
		collSubChaincodeID := args[7]
		repoChannelID := args[8]
		return t.rejectNotification(stub, participantID, notificationType, transactionRef, notificationID, userID, repoChaincodeID, collSubChaincodeID, repoChannelID)
	} else if function == "newMultiParty" {
		return t.newMultiParty(stub, args[1])
	}

	return shim.Error("Received unknown function query invocation with function")
}

func (t *MultiPartyChaincode) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("###### MultiPartyChaincode: function: query ")

	if function == "queryAllNotification" {
		participantID := args[1]
		notificationType := args[2]
		return t.queryAllNotification(stub, participantID, notificationType)
	} else if function == "queryNotification" {
		participantID := args[1]
		notificationType := args[2]
		transactionRef := args[3]
		notificationID := args[4]
		return t.queryNotification(stub, participantID, notificationType, transactionRef, notificationID)
	} else if function == "queryNotificationsByFilter" {
		return t.queryNotificationsByFilter(stub, args[1])
	} else if function == "queryNotificationsByTransactionRef" {
		transactionRef := args[1]
		return t.queryNotificationsByTransactionRef(stub, transactionRef)
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
