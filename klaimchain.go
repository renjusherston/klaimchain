//Author: renju vm

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// KlaimChaincode klaim verification chaincode
type KlaimChaincode struct {
}

var certIndexStr = "_certindex"				//name for the key/value that will store a list of all known certs
var opentransStr = "_opentrans"				//name for the key/value that will store all klaims

type Cert struct{
	Claimant string `json:"claimant"`					//the fieldtags are needed to keep track klaim
	Claimref string `json:"claimref"`
	Policyno string `json:"policyno"`
	Blockrefid string `json:"blockrefid"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Carnumber string `json:"carnumber"`
	Accidentdate string `json:"accidentdate"`
}

type Invoice struct{
	Claimref string `json:"claimref"`
	Invoice string `json:"invoice"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(KlaimChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *KlaimChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("start", []byte(strconv.Itoa(Aval)))				//making a test var "start", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(certIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations
// ============================================================================================================================
func (t *KlaimChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *KlaimChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		res, err := t.Init(stub, "init", args)
		return res, err
	} else if function == "init_cert" {									//create a new  klaim
		res, err := t.init_cert(stub, args)
		return res, err
	} else if function == "init_invoice" {									//create a new  klaim
		res, err := t.init_invoice(stub, args)
		return res, err
	}

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Quering klaims
// ============================================================================================================================
func (t *KlaimChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	} else if function == "readAll" {
		return t.readAll(stub, args)
	} else if function == "validate" {													//read a variable
		return t.validate(stub, args)
	} else if function == "validateinvoice" {													//read a variable
		return t.validateinvoice(stub, args)
	}

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var blockid string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	blockid = strings.ToLower(args[0])

	keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []Cert

		for keysIter.HasNext() {
			key, _, iterErr := keysIter.Next()
			if iterErr != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}
			vals, err := stub.GetState(key)
			if err != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}

			var klaim Cert
			json.Unmarshal(vals, &klaim)

				if(klaim.Blockrefid == blockid){
					keys = append(keys, klaim)
				}

		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
		}

		return jsonKeys, nil

}

// ============================================================================================================================
// Read all - read all matching variable from chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) readAll(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var vecnumber, accdt, refno string

	vecnumber = strings.ToLower(args[0])
	accdt = strings.ToLower(args[1])
	refno = strings.ToLower(args[2])

	keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []Cert
		for keysIter.HasNext() {
			key, _, iterErr := keysIter.Next()
			if iterErr != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}
			vals, err := stub.GetState(key)
			if err != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}

			var klaim Cert
			json.Unmarshal(vals, &klaim)

			if(accdt != ""){
				if(klaim.Carnumber == vecnumber && klaim.Accidentdate == accdt){
					keys = append(keys, klaim)
				}
			}else{
				if(klaim.Claimref == refno){
					keys = append(keys, klaim)
				}
			}

		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
		}

		return jsonKeys, nil


}


// ============================================================================================================================
// Validate - validate a variable from chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) validate(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var vecnumber, accdt, email string

	vecnumber = strings.ToLower(args[0])
	accdt = strings.ToLower(args[1])
	email = strings.ToLower(args[2])

	keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []Cert

		for keysIter.HasNext() {
			key, _, iterErr := keysIter.Next()
			if iterErr != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}
			vals, err := stub.GetState(key)
			if err != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}

			var klaim Cert
			json.Unmarshal(vals, &klaim)

			if(klaim.Carnumber == vecnumber && klaim.Accidentdate == accdt && klaim.Email == email){
				keys = append(keys, klaim)
			}else if(klaim.Carnumber == vecnumber && klaim.Accidentdate == accdt && klaim.Email != email){
				keys = append(keys, klaim)
			}
		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
		}

		return jsonKeys, nil

}

// ============================================================================================================================
// Validate - validate invoice from chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) validateinvoice(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var invoice string

	invoice = strings.ToLower(args[0])


	keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []Invoice

		for keysIter.HasNext() {
			key, _, iterErr := keysIter.Next()
			if iterErr != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}
			vals, err := stub.GetState(key)
			if err != nil {
				return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
			}

			var klaim Invoice
			json.Unmarshal(vals, &klaim)

				if(klaim.Invoice == invoice){
					keys = append(keys, klaim)
				}

		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
		}

		return jsonKeys, nil

}

// ============================================================================================================================
// Init cert - create a new klaim, store into chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) init_cert(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	ctime := time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))

	claimref := strings.ToLower(args[0])
	blockrefid := strings.ToLower(args[1])
	claimant := strings.ToLower(args[2])
	phone := strings.ToLower(args[3])
	email := strings.ToLower(args[4])
	carnumber := strings.ToLower(args[5])
	accidentdate := strings.ToLower(args[6])
	policyno := strings.ToLower(args[7])

	//build the cert json string manually
	str := `{"claimref": "` + claimref + `", "policyno": "` + policyno + `", "blockrefid": "` + blockrefid + `", "claimant": "` + claimant + `", "phone": "` + phone + `", "email": "` + email + `", "carnumber": "` + carnumber + `", "accidentdate": "` + accidentdate + `"}`

	err = stub.PutState(strconv.FormatInt(ctime,10), []byte(str))									//store cert with user name as key
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Init invoice - create a invoice entry, store into chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) init_invoice(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	ctime := time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))

	claimref := strings.ToLower(args[0])
	invoice := strings.ToLower(args[1])

	//build the cert json string manually
	str := `{"claimref": "` + claimref + `", "invoice": "` + invoice + `"}`

	err = stub.PutState(strconv.FormatInt(ctime,10), []byte(str))									//store cert with user name as key
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}
