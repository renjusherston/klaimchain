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
	Insuarer string `json:"insuarer"`					//the fieldtags are needed to keep track klaim
	Klaimdate string `json:"klaimdate"`
	Doctype string `json:"doctype"`
	Dochash string `json:"dochash"`
}

type AnOpenCert struct{
	Insuarer string `json:"insuarer_name"`					//user who created the Klaim
	Dochash string `json:"doc_hash"`
	Timestamp int64 `json:"timestamp"`			//utc timestamp of creation
}

type AllTransactions struct{
	opentrans []AnOpenCert `json:"open_Transactions"`
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

	var Transactions AllTransactions
	jsonAsBytes, _ = json.Marshal(Transactions)								//clear the open transaction struct
	err = stub.PutState(opentransStr, jsonAsBytes)
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
	}

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var doc string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	doc = strings.ToLower(args[0])

	keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []string
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

				if(klaim.Dochash == doc){
					keys = append(keys, klaim.Insuarer+","+klaim.Klaimdate+","+klaim.Doctype+","+klaim.Dochash)
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

	var name, dt string

	name = strings.ToLower(args[0])
	dt = strings.ToLower(args[1])

	keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []string
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

			if(dt != ""){
				if(klaim.Insuarer == name && klaim.Klaimdate == dt){
					keys = append(keys, klaim.Insuarer+","+klaim.Klaimdate+","+klaim.Doctype+","+klaim.Dochash)
				}
			}else{
				if(klaim.Insuarer == name){
					keys = append(keys, klaim.Insuarer+","+klaim.Klaimdate+","+klaim.Doctype+","+klaim.Dochash)
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
// Init cert - create a new klaim, store into chaincode state
// ============================================================================================================================
func (t *KlaimChaincode) init_cert(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	ctime := time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))


	insuarer := strings.ToLower(args[0])
	klaimdate := strings.ToLower(args[1])
	doctype := strings.ToLower(args[2])
	dochash := strings.ToLower(args[3])

	//build the cert json string manually
	str := `{"insuarer": "` + insuarer + `", "klaimdate": "` + klaimdate + `", "doctype": "` + doctype + `", "dochash": "` + dochash + `"}`

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
