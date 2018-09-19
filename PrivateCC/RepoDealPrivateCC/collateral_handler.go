package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type CollateralDetails struct {
	ObjectType                   string `json:"ObjectType,omitempty"`
	ProcessingSystemReference    string `json:"ProcessingSystemReference,omitempty"`
	CollateralType	             string `json:"CollateralType,omitempty"` //PLEDGE, BORROW, REPO
	LenderParticipantID          string `json:"LenderParticipantID,omitempty"`   //Lender of Collateral
	LenderParticipantAcct		 string `json:"LenderParticipantAcct,omitempty"`   //Lender of Collateral
	BorrowerParticipantID        string `json:"BorrowerParticipantID,omitempty"` //Borrower of Collareral
	BorrowerParticipantAcct      string `json:"BorrowerParticipantAcct,omitempty"` //Borrower of Collareral
	Instrument                   string `json:"Instrument,omitempty"`
	CusipDescription             string `json:"CusipDescription,omitempty"`
	AssetClass                   string `json:"AssetClass,omitempty"`
	SubAccount                   string `json:"SubAccount,omitempty"`
	TransactionDate              string `json:"TransactionDate,omitempty"`
	TransactionTimestamp         string `json:"TransactionTimestamp,omitempty"`
	EffectiveDate                string `json:"EffectiveDate,omitempty"`        //Same as TransactionDate
	ContractualValueDate         string `json:"ContractualValueDate,omitempty"`
	CloseEventDate               string `json:"CloseEventDate,omitempty"`
	Quantity                     string `json:"Quantity,omitempty"`
	CleanPrice                   string `json:"CleanPrice,omitempty"`
	DirtyPrice                   string `json:"DirtyPrice,omitempty"`
	Principal                    string `json:"Principal,omitempty"`
	Haircut                      string `json:"Haircut,omitempty"`
	AccruedInterestNoOfDays      int    `json:"AccruedInterestNoOfDays,string,omitempty"` // COUPON ACC DAYSS
	CouponAccruedInterest        string `json:"CouponAccruedInterest,omitempty"`
	Factor                       int    `json:"Factor,string,omitempty"`
	NetConsiderationBaseCurrency string `json:"NetConsiderationBaseCurrency,omitempty"`
	CurrentQuantity              string `json:"CurrentQuantity,omitempty"`
	CurrentPrice                 string `json:"CurrentPrice,omitempty"`
	CurrentValue                 string `json:"CurrentValue,omitempty"`
	LastUpdatedUser              string `json:"LastUpdatedUser,omitempty"`
	DateTime                     string `json:"DateTime,omitempty"`
	EditFlag                     string `json:"EditFlag,omitempty"`
	Version                      int    `json:"Version,string,omitempty"` // UPDATED BY CHAINCODE
	ActiveInd                    string `json:"ActiveInd,omitempty"`      // UPDATED BY CHAINCODE
}

//CollateralHandler provides APIs used to perform operations on CC's KV store
type collateralHandler struct {
}

// NewCollateralHandler create a new reference to CertHandler
func NewCollateralHandler() *collateralHandler {
	return &collateralHandler{}
}

var utilHandler = NewUtilityHandler()

// newCollateralPosition adds the record row on the chaincode state
func (t *collateralHandler) newCollateralPosition(stub shim.ChaincodeStubInterface, collStruct CollateralDetails) error {

	fmt.Println("###### RepoDealCC: function: newCollateralPosition ")
	var err error
	collStruct.ObjectType = "CollateralDetails"
	collStruct.ActiveInd = "A"
	collStruct.EditFlag = "X"
	collStruct.Version = 1
	collStruct.DateTime = time.Now().UTC().String()

	compositeKey, _ := stub.CreateCompositeKey(collStruct.ObjectType, []string{collStruct.BorrowerParticipantID, collStruct.ProcessingSystemReference, collStruct.Instrument, strconv.Itoa(collStruct.Version)})
	collJSONBytes, _ := json.Marshal(collStruct)

	fmt.Println("Collateral CompositeKey::", compositeKey)
	err = stub.PutPrivateData(collection, compositeKey, collJSONBytes)
	if err != nil {
		return errors.New("Error in adding Collateral state")
	}

	return nil
}

// updateCollateralPosition replaces the collateral record row on the chaincode state
func (t *collateralHandler) updateCollateralPosition(stub shim.ChaincodeStubInterface, collStruct CollateralDetails) error {

	var err error
	fmt.Println("###### RepoDealCC: function: updateCollateralPosition ")

	collStruct.ObjectType = "CollateralDetails"

	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(collStruct.ObjectType, []string{collStruct.BorrowerParticipantID, collStruct.ProcessingSystemReference, collStruct.Instrument, strconv.Itoa(collStruct.Version)})
	extJSONBytes, _ := stub.GetState(compositeKey)
	fmt.Printf("Existing Collateral", string(compositeKey), string(extJSONBytes))

	var exCollStruct CollateralDetails
	if string(extJSONBytes) != "" {
		err = json.Unmarshal([]byte(extJSONBytes), &exCollStruct)
		if err != nil {
			return errors.New("Error unmarshalling structure")
		}

		exCollStruct.ActiveInd = "N"
		exCollStruct.DateTime = time.Now().UTC().String()
		extJSONBytes, _ = json.Marshal(exCollStruct)
		err = stub.PutPrivateData(collection, compositeKey, extJSONBytes)
		if err != nil {
			return errors.New("Error in updating Collateral state")
		}
	}
	//Create new version and document
	collStruct.ActiveInd = "A"
	collStruct.EditFlag = "X"
	collStruct.Version = collStruct.Version + 1
	collStruct.DateTime = time.Now().UTC().String()

	compositeKey, _ = stub.CreateCompositeKey(collStruct.ObjectType, []string{collStruct.BorrowerParticipantID, collStruct.ProcessingSystemReference, collStruct.Instrument, strconv.Itoa(collStruct.Version)})
	collJSONBytes, _ := json.Marshal(collStruct)
	fmt.Printf("New Collateral", string(compositeKey), string(collJSONBytes))

	err = stub.PutPrivateData(collection, compositeKey, collJSONBytes)
	if err != nil {
		return errors.New("Error in adding Collateral state")
	}

	return nil
}

// deactivateCollateralPosition deactivate the collateral record row on the chaincode state
func (t *collateralHandler) deactivateCollateralPosition(stub shim.ChaincodeStubInterface, collStruct CollateralDetails) error {

	fmt.Println("###### RepoDealCC: function: deactivateCollateralPosition ")

	collStruct.ObjectType = "CollateralDetails"
	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(collStruct.ObjectType, []string{collStruct.BorrowerParticipantID, collStruct.ProcessingSystemReference, collStruct.Instrument, strconv.Itoa(collStruct.Version)})
	extJSONBytes, _ := stub.GetState(compositeKey)

	var exCollStruct CollateralDetails
	if string(extJSONBytes) != "" {
		err := json.Unmarshal([]byte(extJSONBytes), &exCollStruct)
		if err != nil {
			return errors.New("Error unmarshalling structure")
		}

		exCollStruct.ActiveInd = "N"
		exCollStruct.DateTime = time.Now().UTC().String()
		extJSONBytes, _ = json.Marshal(exCollStruct)
		err = stub.PutPrivateData(collection, compositeKey, extJSONBytes)
		if err != nil {
			return errors.New("Error in deactivating Collateral state")
		}
	}
	return nil
}

func (t *collateralHandler) newCollateralCapture(stub shim.ChaincodeStubInterface, collStruct CollateralDetails, action string) error {

	fmt.Println("###### RepoDealCC: function: newCollateralCapture ")
	var err error
	if action == "NEW" {
		err = t.newCollateralPosition(stub, collStruct)
		if err != nil {
			fmt.Println("Error adding new collateral [%v]", err)
		}

	} else if action == "AMEND" {
		err = t.updateCollateralPosition(stub, collStruct)
		if err != nil {
			fmt.Println("Error adding new collateral [%v]", err)
		}

	} else if action == "CANCEL" {
		err = t.deactivateCollateralPosition(stub, collStruct)
		if err != nil {
			fmt.Println("Error adding new collateral [%v]", err)
		}

	}

	return nil
}

func (t *collateralHandler) newCollateralValUpdate(stub shim.ChaincodeStubInterface, collStruct CollateralDetails, action string) error {

	fmt.Println("###### RepoDealCC: function: newCollateralValUpdate ")
	var err error
	if action == "NEW" {
		err = t.newCollateralPosition(stub, collStruct)
		if err != nil {
			return err
		}

	} else if action == "AMEND" {
		err = t.updateCollateralPosition(stub, collStruct)
		if err != nil {
			return err
		}

	} else if action == "CANCEL" {
		err = t.deactivateCollateralPosition(stub, collStruct)
		if err != nil {
			return err
		}

	}

	return nil
}

// queryCollateralPosition returns the record row matching a corresponding position on the chaincode state
func (t *collateralHandler) queryCollateralPosition(stub shim.ChaincodeStubInterface, borrowerParticipantID string, processingSystemReference string, instrumentID string, version string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryCollateralPosition ")

	if borrowerParticipantID != "" && processingSystemReference != "" && instrumentID != "" && version != "" {

		var attributes []string
		attributes = append(attributes, borrowerParticipantID)
		attributes = append(attributes, processingSystemReference)
		attributes = append(attributes, instrumentID)
		attributes = append(attributes, version)

		collJSONBytes, err := utilHandler.readSingleJSON(stub, "CollateralDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving Collateral")
		}
		return collJSONBytes, nil
	}

	return nil, nil
}

// queryCollateralPositionByInstrument returns the record row matching a corresponding position on the chaincode state
func (t *collateralHandler) queryCollateralPositionByInstrument(stub shim.ChaincodeStubInterface, borrowerParticipantID string, processingSystemReference string, instrumentID string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryCollateralPositionByInstrument ")

	if borrowerParticipantID != "" && processingSystemReference != "" && instrumentID != "" {

		var attributes []string
		attributes = append(attributes, borrowerParticipantID)
		attributes = append(attributes, processingSystemReference)
		attributes = append(attributes, instrumentID)

		collJSONBytes, err := utilHandler.readMultiCollateralJSON(stub, "CollateralDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving Collateral")
		}
		return collJSONBytes, nil
	}

	return nil, nil
}

// queryAllCollateralPositionsByRepo that returns the active record row matching a correponding Participant ID and System reference on the chaincode state
func (t *collateralHandler) queryAllCollateralPositionsByRepo(stub shim.ChaincodeStubInterface, borrowerParticipantID string, processingSystemReference string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryAllCollateralPositionsByRepo ")
	if borrowerParticipantID != "" && processingSystemReference != "" {

		var attributes []string
		attributes = append(attributes, borrowerParticipantID, processingSystemReference)
		fmt.Println("Collateral Info::", borrowerParticipantID, processingSystemReference)
		finaldata, err := utilHandler.readMultiCollateralJSON(stub, "CollateralDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving multi Collateral")
		}

		var finaldata1 []byte
		prefix := "{\"Collateral\" : ["
		finaldata1 = append(finaldata1, prefix...)
		finaldata1 = append(finaldata1, finaldata...)
		suffix := "]}"
		finaldata1 = append(finaldata1, suffix...)

		return []byte(finaldata1), nil
	}

	return nil, nil
}

// queryAllCollateralPositionsByParticipant that returns the active record row matching a correponding Participant ID and System reference on the chaincode state
func (t *collateralHandler) queryAllCollateralPositionsByParticipant(stub shim.ChaincodeStubInterface, borrowerParticipantID string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryAllCollateralPositionsByParticipant ")
	if borrowerParticipantID != "" {

		var attributes []string
		attributes = append(attributes, borrowerParticipantID)
		fmt.Println("Collateral Info::", borrowerParticipantID)
		finaldata, err := utilHandler.readMultiCollateralJSONByParticipant(stub, "CollateralDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving multi Collateral")
		}

		if finaldata != nil {
			var finaldata1 []byte
			prefix := "{\"Collateral\" : ["
			finaldata1 = append(finaldata1, prefix...)
			finaldata1 = append(finaldata1, finaldata...)
			suffix := "]}"
			finaldata1 = append(finaldata1, suffix...)

			return []byte(finaldata1), nil
		}
	}

	return nil, nil
}
