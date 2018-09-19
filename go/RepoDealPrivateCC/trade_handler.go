package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type TradeDetails struct {
	ObjectType                string `json:"ObjectType,omitempty"`
	ProcessingSystemReference string `json:"ProcessingSystemReference,omitempty"`
	FOReference               string `json:"FOReference,omitempty"`
	ExternalReference         string `json:"ExternalReference,omitempty"`
	TranOriginatorParty       string `json:"TranOriginatorParty,omitempty"`   //SET WHOEVER INITATE THE TRANSACTION
	Party                     string `json:"Party,omitempty"`                 // PARTICIPANT_ID ON THE PARTY JSON
	PartyCustodian            string `json:"PartyCustodian,omitempty"`        // PARTICIPANT_ID ON THE PARTY CUSTODIAN JSON
	Counterparty              string `json:"Counterparty,omitempty"`          // PARTICIPANT_ID ON THE COUNTERPARTY JSON
	CounterpartyCustodian     string `json:"CounterpartyCustodian,omitempty"` // PARTICIPANT_ID ON THE COUNTERPARTY CUSTODIAN JSONs
	TradeType                 string `json:"TradeType,omitempty"`             // BORROW VS LOAN, REPO etc.
	TransactionType           string `json:"TransactionType,omitempty"`       // FICC, NON-FICC, INTERNAL
	TransactionStatus         string `json:"TransactionStatus,omitempty"`     // NEW, AMEND, CANCEL
	RepoStatus                string `json:"RepoStatus,omitempty"`            // PENDAPPROVAL, APPROVED, REJECTED--REPO DEAL CC STATUS: OPENLEGPEND, OPENLEGSETTLED, CLOSELEGPEND, CLOSELEGSETTLED, COLLSUBPEND, INTPAYMENTPEND
	TransactionDate           string `json:"TransactionDate,omitempty"`
	TransactionTimestamp      string `json:"TransactionTimestamp,omitempty"`
	EffectiveDate             string `json:"EffectiveDate,omitempty"`
	ContractualValueDate      string `json:"ContractualValueDate,omitempty"`
	CloseEventDate            string `json:"CloseEventDate,omitempty"`
	PlaceOfTrade              string `json:"PlaceOfTrade,omitempty"`
	BaseCurrency              string `json:"BaseCurrency,omitempty"`
	SettleCurrency            string `json:"SettleCurrency,omitempty"`
	TotalCashAmount           string `json:"TotalCashAmount,omitempty"` // CASH AMOUNT = SUM OF ALL COLLATERAL CASH AMT
	TotalPrincipalAmount      string `json:"TotalPrincipalAmount,omitempty"` 
	TotalNetAmount            string `json:"TotalNetAmount,omitempty"` 
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

//TradeHandler provides APIs used to perform operations on CC's KV store
type tradeHandler struct {
}

// NewTradeHandler create a new reference to CertHandler
func NewTradeHandler() *tradeHandler {
	return &tradeHandler{}
}

// newTradeEntry adds the trade record on the chaincode state table
func (t *tradeHandler) newTradeEntry(stub shim.ChaincodeStubInterface, tradeStruct TradeDetails) error {

	fmt.Println("###### RepoDealCC: function: newTradeEntry ")
	tradeStruct.ObjectType = "TradeDetails"
	tradeStruct.TransactionStatus = "NEW"
	tradeStruct.AccruedInterest = "0.00"
	tradeStruct.TotalPaidInterest = "0.00"
	tradeStruct.Version = 1
	tradeStruct.ActiveInd = "A"
	tradeStruct.DateTime = time.Now().UTC().String()
	//tradeStruct.RepoStatus = "PENDAPPROVAL"

	compositeKey, err := stub.CreateCompositeKey(tradeStruct.ObjectType, []string{tradeStruct.ProcessingSystemReference, strconv.Itoa(tradeStruct.Version)})
	tradeJSONBytes, err := json.Marshal(tradeStruct)
	collection := "RepoDealCollection"

	err = stub.PutPrivateData(collection,compositeKey, tradeJSONBytes)
	if err != nil {
		return errors.New("Error in adding trade state")
	}

	return nil
}

// updateTrade replaces the trade record row on the chaincode state table
func (t *tradeHandler) updateTrade(stub shim.ChaincodeStubInterface, tradeStruct TradeDetails) error {
	fmt.Println("###### RepoDealCC: function: updateTrade ")

	tradeStruct.ObjectType = "TradeDetails"
	fmt.Println("Existing Trade Input : ", tradeStruct.ProcessingSystemReference, tradeStruct.Version)
	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(tradeStruct.ObjectType, []string{tradeStruct.ProcessingSystemReference, strconv.Itoa(tradeStruct.Version)})
	extJSONBytes, _ := stub.GetState(compositeKey)
	var exTradeStruct TradeDetails
	if string(extJSONBytes) != "" {
		err := json.Unmarshal([]byte(extJSONBytes), &exTradeStruct)
		if err != nil {
			fmt.Println("Error parsing trade JSON [%v]", err)
			return err
		}
		exTradeStruct.ActiveInd = "N"
		exTradeStruct.DateTime = time.Now().UTC().String()
		extJSONBytesNew, _ := json.Marshal(exTradeStruct)
		collection := "RepoDealCollection"
		fmt.Println("Existing Trade : ", string(compositeKey), string(extJSONBytesNew))
		err = stub.PutPrivateData(collection,compositeKey, extJSONBytesNew)
		if err != nil {
			return errors.New("Error in updating trade state")
		}

	}

	// Create a new version and document
	tradeStruct.Version = tradeStruct.Version + 1
	tradeStruct.ActiveInd = "A"
	tradeStruct.TransactionStatus = "AMEND"
	//tradeStruct.RepoStatus = "PENDAPPROVAL"
	tradeStruct.DateTime = time.Now().UTC().String()
	compositeKey, _ = stub.CreateCompositeKey(tradeStruct.ObjectType, []string{tradeStruct.ProcessingSystemReference, strconv.Itoa(tradeStruct.Version)})
	tradeJSONBytes, _ := json.Marshal(tradeStruct)
	collection := "RepoDealCollection"
	fmt.Println("New Trade : ", string(compositeKey), string(tradeJSONBytes))

	err := stub.PutPrivateData(collection,compositeKey, tradeJSONBytes)
	if err != nil {
		return errors.New("Error in adding trade state")
	}

	return nil
}

// deactivateTrade replaces the trade record row on the chaincode state table
func (t *tradeHandler) deactivateTrade(stub shim.ChaincodeStubInterface, tradeStruct TradeDetails) error {

	var err error
	fmt.Println("###### RepoDealCC: function: deactivateTrade ")
	
	tradeStruct.ObjectType = "TradeDetails"
	fmt.Println("Existing Trade Input : ", tradeStruct.ProcessingSystemReference, tradeStruct.Version)
	//Get Version Number from Query function. Expected to be set in the input.
	compositeKey, _ := stub.CreateCompositeKey(tradeStruct.ObjectType, []string{tradeStruct.ProcessingSystemReference, strconv.Itoa(tradeStruct.Version)})
	extJSONBytes, _ := stub.GetState(compositeKey)
	var exTradeStruct TradeDetails
	if string(extJSONBytes) != "" {
		err = json.Unmarshal([]byte(extJSONBytes), &exTradeStruct)
		if err != nil {
			fmt.Println("Error parsing trade JSON [%v]", err)
			return err
		}
		exTradeStruct.ActiveInd = "N"
		exTradeStruct.DateTime = time.Now().UTC().String()
		extJSONBytesNew, _ := json.Marshal(exTradeStruct)
		collection := "RepoDealCollection"
		fmt.Println("Existing Trade : ", string(compositeKey), string(extJSONBytesNew))
		err = stub.PutPrivateData(collection,compositeKey, extJSONBytesNew)
		if err != nil {
			return errors.New("Error in updating trade state")
		}

	}

	// Create a new version and document
	tradeStruct.Version = tradeStruct.Version + 1
	tradeStruct.ActiveInd = "A"
	tradeStruct.TransactionStatus = "CANCEL"
	//tradeStruct.RepoStatus = "PENDAPPROVAL"
	tradeStruct.DateTime = time.Now().UTC().String()
	compositeKey, _ = stub.CreateCompositeKey(tradeStruct.ObjectType, []string{tradeStruct.ProcessingSystemReference, strconv.Itoa(tradeStruct.Version)})
	tradeJSONBytes, _ := json.Marshal(tradeStruct)
	collection := "RepoDealCollection"
	fmt.Println("New Trade : ", string(compositeKey), string(tradeJSONBytes))

	err = stub.PutPrivateData(collection,compositeKey, tradeJSONBytes)
	if err != nil {
		return errors.New("Error in adding trade state")
	}

	return nil
}

func (t *tradeHandler) newTradeCapture(stub shim.ChaincodeStubInterface, tradeStruct TradeDetails, action string) error {

	fmt.Println("###### RepoDealCC: function: newTradeCapture ")
	var err error
	if action == "NEW" {
		err = t.newTradeEntry(stub, tradeStruct)
		if err != nil {
			fmt.Println("Error adding new trade [%v]", err)
		}

	} else if action == "AMEND" {
		err = t.updateTrade(stub, tradeStruct)
		if err != nil {
			fmt.Println("Error updating new trade [%v]", err)
		}

	} else if action == "CANCEL" {
		err = t.deactivateTrade(stub, tradeStruct)
		if err != nil {
			fmt.Println("Error deactivate new trade [%v]", err)
		}

	}

	return nil
}

func (t *tradeHandler) interestCalculation(stub shim.ChaincodeStubInterface, tradeStruct TradeDetails) error {

	fmt.Println("###### RepoDealCC: function: interestCalculation ")
	//Calculate Daily Interest
	var interest float64
	var finRate float64
	var cashAmt float64
	var currentInterest float64
	var err error
	finRate, _ = strconv.ParseFloat(tradeStruct.CurrentFinancingRate, 64)
	cashAmt, _ = strconv.ParseFloat(tradeStruct.TotalCashAmount, 64)
	currentInterest, _ = strconv.ParseFloat(tradeStruct.AccruedInterest, 64)
	fmt.Println("Values : ", finRate, cashAmt, currentInterest)

	interest = ((finRate / 100) * cashAmt) / 360 
	fmt.Println("New interest calculation : ", interest)
	currentInterest = currentInterest + interest
	fmt.Println("New Current interest calculation : ", currentInterest)

	tradeStruct.AccruedInterest = strconv.FormatFloat(currentInterest, 'f', 2, 64)
	tradeStruct.InterestDays = tradeStruct.InterestDays + 1
	tradeStruct.LastUpdatedUser = "InterestCalService"
	err = t.updateTrade(stub, tradeStruct)
	if err != nil {
		return errors.New("Error in calculating interest")
	}
	return nil
}

func (t *tradeHandler) interestPayment(stub shim.ChaincodeStubInterface, tradeStruct TradeDetails, payment string) error {

	fmt.Println("###### RepoDealCC: function: interestPayment ")
	//Calculate Daily Interest
	var paymentFloat float64
	var totalInterest float64
	var accInt float64
	var err error
	fmt.Println("Payment Amount: ", payment)
	fmt.Println("Current interest Paid : ", tradeStruct.TotalPaidInterest)
	fmt.Println("Currrent Accrued Interest Paid : ", tradeStruct.AccruedInterest)

	paymentFloat, _ = strconv.ParseFloat(payment, 64)
	totalInterest, _ = strconv.ParseFloat(tradeStruct.TotalPaidInterest, 64)
	accInt, _ = strconv.ParseFloat(tradeStruct.AccruedInterest, 64)

	if paymentFloat <= accInt {
		totalInterest = totalInterest + paymentFloat
		accInt = accInt - paymentFloat
		tradeStruct.LastUpdatedUser = "InterestPaymentService"
		tradeStruct.TotalPaidInterest = strconv.FormatFloat(totalInterest, 'f', 2, 64)
		tradeStruct.AccruedInterest = strconv.FormatFloat(accInt, 'f', 2, 64)
		fmt.Println("New interest Paid : ", tradeStruct.TotalPaidInterest)
		fmt.Println("New Accrued Interest Paid : ", tradeStruct.AccruedInterest)
		tradeStruct.RepoStatus = "OPENLEGSETTLED"

		err = t.updateTrade(stub, tradeStruct)
		if err != nil {
			return errors.New("Error in calculating interest")
		}
	} else {
		err = repHandler.repoStatusUpdate(stub, tradeStruct.ProcessingSystemReference, "OPENLEGSETTLED", "RepoDeal Service")
		return errors.New("Not enough interest amount")
	}
	return nil
}

func (t *tradeHandler) cashAdjustmentPayment(stub shim.ChaincodeStubInterface, tradeStruct TradeDetails, payment string, indicator string) error {

	fmt.Println("###### RepoDealCC: function: cashAdjustmentPayment ")
	//Calculate Daily Interest
	var paymentFloat float64
	var totalCashAmount float64
	var err error
	fmt.Println("Cash Payment Amount: ", payment, indicator)
	fmt.Println("Current cash amount : ", tradeStruct.TotalCashAmount)

	paymentFloat, _ = strconv.ParseFloat(payment, 64)
	totalCashAmount, _ = strconv.ParseFloat(tradeStruct.TotalCashAmount, 64)

	if indicator == "DEBIT" {
		if paymentFloat <= totalCashAmount {
			totalCashAmount = totalCashAmount - paymentFloat
			tradeStruct.LastUpdatedUser = "CashAdjustmentService"
			tradeStruct.TotalCashAmount = strconv.FormatFloat(totalCashAmount, 'f', 2, 64)
			fmt.Println("New Cash Amount : ", tradeStruct.TotalCashAmount)

			tradeStruct.RepoStatus = "OPENLEGSETTLED"

			err = t.updateTrade(stub, tradeStruct)
			if err != nil {
				return errors.New("Error in updating trade")
			}
		} else {
			err = repHandler.repoStatusUpdate(stub, tradeStruct.ProcessingSystemReference, "OPENLEGSETTLED", "RepoDeal Service")
			return errors.New("Not enough interest amount")
		}
	} else {
		totalCashAmount = totalCashAmount + paymentFloat
		tradeStruct.LastUpdatedUser = "CashAdjustmentService"
		tradeStruct.TotalCashAmount = strconv.FormatFloat(totalCashAmount, 'f', 2, 64)
		fmt.Println("New Cash Amount : ", tradeStruct.TotalCashAmount)

		tradeStruct.RepoStatus = "OPENLEGSETTLED"

		err = t.updateTrade(stub, tradeStruct)
		if err != nil {
			return errors.New("Error in updating trade")
		}
	}
	return nil
}

// queryTrade returns the record row matching a corresponding trade on the chaincode state table
func (t *tradeHandler) queryTrade(stub shim.ChaincodeStubInterface, tradeRef string, version string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryTrade ")
	if tradeRef != "" {

		var attributes []string
		attributes = append(attributes, tradeRef)
		attributes = append(attributes, version)

		tradeJSONBytes, err := utilHandler.readSingleJSON(stub, "TradeDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving trade")

		}
		return tradeJSONBytes, nil
	}

	return nil, nil
}

// queryActiveTrade returns the active record row matching a corresponding trade on the chaincode state
func (t *tradeHandler) queryActiveTrade(stub shim.ChaincodeStubInterface, tradeRef string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryActiveTrade ")
	if tradeRef != "" {

		var attributes []string
		attributes = append(attributes, tradeRef)

		tradeJSONBytes, err := utilHandler.readTradeJSON(stub, "TradeDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving trade")

		}
		return tradeJSONBytes, nil
	}

	return nil, nil
}

// queryAllActiveTrade returns the active record row matching a corresponding trade on the chaincode state
func (t *tradeHandler) queryAllActiveTrade(stub shim.ChaincodeStubInterface) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryAllActiveTrade ")

	var attributes []string
	//attributes = append(attributes, tradeRef)

	tradeJSONBytes, err := utilHandler.readMultiTradeActiveJSON(stub, "TradeDetails", attributes)
	if err != nil {
		return nil, errors.New("Error retriving trade")

	}

	if tradeJSONBytes != nil {
		var finaldata1 []byte
		prefix := "{\"Trade\" : ["
		finaldata1 = append(finaldata1, prefix...)
		finaldata1 = append(finaldata1, tradeJSONBytes...)
		suffix := "]}"
		finaldata1 = append(finaldata1, suffix...)

		return []byte(finaldata1), nil
	}
	/*queryString := fmt.Sprintf("{\"selector\":{\"ObjectType\":\"TradeDetails\",\"ActiveInd\":\"A\"}}")
	queryResults, err := utilHandler.getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return nil, err
	}*/

	//return []byte(queryResults), nil
	return nil, nil
}

// queryActiveTrade returns the active record row matching a corresponding trade on the chaincode state
func (t *tradeHandler) queryTradeHistoryReport(stub shim.ChaincodeStubInterface, processingSystemReference string) ([]byte, error) {

	fmt.Println("###### RepoDealCC: function: queryTradeHistoryReport ")
	if processingSystemReference != "" {

		var attributes []string
		attributes = append(attributes, processingSystemReference)
		fmt.Println("Trade Info::", processingSystemReference)
		finaldata, err := utilHandler.readMultiTradeJSON(stub, "TradeDetails", attributes)
		if err != nil {
			return nil, errors.New("Error retriving multi trade")
		}

		var finaldata1 []byte
		prefix := "{\"TradeHistory\" : ["
		finaldata1 = append(finaldata1, prefix...)
		finaldata1 = append(finaldata1, finaldata...)
		suffix := "]}"
		finaldata1 = append(finaldata1, suffix...)

		return []byte(finaldata1), nil
	}

	return nil, nil
}
