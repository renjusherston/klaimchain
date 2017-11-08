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

// CertificateChaincode certificate verification chaincode
type CertificateChaincode struct {
}

var certIndexStr = "_certindex"				//name for the key/value that will store a list of all known certs
var opentransStr = "_opentrans"				//name for the key/value that will store all certificates

type Cert struct{
	Owner string `json:"owner_name"`					//the fieldtags are needed to keep track certificate
	Unittitle string `json:"unit_title"`
	Qualid string `json:"qual_identifier"`
	Unitid string `json:"unit_identifier"`
  User string `json:"user_name"`
  Certificate string `json:"cert_hash"`
}

type Description struct{
	Unittitle string `json:"unit_title"`
	Qualid string `json:"qual_identifier"`
  Certificate string `json:"cert_hash"`
}

type AnOpenCert struct{
	Owner string `json:"owner_name"`					//user who created the Certificate
	Timestamp int64 `json:"timestamp"`			//utc timestamp of creation
	Want Description  `json:"want"`				//description of desired certificate
	Willing []Description `json:"willing"`		//array of cert willing to generate certificate
}

type AllTransactions struct{
	opentrans []AnOpenCert `json:"open_Transactions"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(CertificateChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *CertificateChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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
func (t *CertificateChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *CertificateChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	}  else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_cert" {									//create a new  certificate
		return t.init_cert(stub, args)
	} else if function == "set_user" {										//change owner of a certificate
		res, err := t.set_user(stub, args)
		return res, err
	}

	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Quering certificates
// ============================================================================================================================
func (t *CertificateChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query certificate is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")

/*
	var err error

	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting at least 1")
	}

	//get the cert index
	certsAsBytes, err := stub.GetState(certIndexStr)

	if err != nil {
		return nil, errors.New("Failed to get certificate index")
	}

	var certIndex []string
	json.Unmarshal(certsAsBytes, &certIndex)

	for i:= range certIndex{

		certAsBytes, err := stub.GetState(certIndex[i])						//grab this cert
		if err != nil {
			return nil, errors.New("Failed to get Certificate")
		}
		res := Cert{}
		json.Unmarshal(certAsBytes, &res)



		//check for user && certificate
		if strings.ToLower(res.Certificate) == strings.ToLower(args[0]) || strings.ToLower(res.User) == strings.ToLower(args[0]){
			fmt.Println("found a Certificate issued by: " + res.Owner)
			fmt.Println("! end find Certificate")

			return certAsBytes, nil

		}
	}
	if err != nil {
		return nil, err
	}

	fmt.Println("- end find Certificate - error")

	return nil, nil
	*/
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *CertificateChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *CertificateChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init cert - create a new certificate, store into chaincode state
// ============================================================================================================================
func (t *CertificateChaincode) init_cert(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error


	if len(args) != 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}

	//input sanitation
	fmt.Println("- start init cert")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return nil, errors.New("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return nil, errors.New("6th argument must be a non-empty string")
	}


	owner := strings.ToLower(args[0])
	unittitle := strings.ToLower(args[1])
	qualid := strings.ToLower(args[2])
	unitid := strings.ToLower(args[3])
	user := strings.ToLower(args[4])
	certificate := strings.ToLower(args[5])


	//check if cert already exists
	certAsBytes, err := stub.GetState(certificate)
	if err != nil {
		return nil, errors.New("Failed to get certificate")
	}
	res := Cert{}
	json.Unmarshal(certAsBytes, &res)
	if res.Certificate == certificate{
		fmt.Println("This certificate arleady exists: " + certificate)
		fmt.Println(res);
		return nil, errors.New("This certificate arleady exists")				//all stop a cert by this name exists
	}

	//build the cert json string manually
	str := `{"owner": "` + owner + `", "unittitle": "` + unittitle + `", "qualid": "` + qualid + `", "unitid": "` + unitid + `", "user": "` + user + `", "certificate": "` + certificate + `"}`
	err = stub.PutState(user, []byte(str))									//store cert with user name as key

	if err != nil {
		return nil, err
	}

  err = stub.PutState(certificate, []byte(str))									//store  with cert as key

	//get the cert index
	certsAsBytes, err := stub.GetState(certIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get cert index")
	}
	var certIndex []string
	json.Unmarshal(certsAsBytes, &certIndex)							// JSON.parse()

	//append
	certIndex = append(certIndex, user)									//add cert name to index list
	fmt.Println("! cert index: ", certIndex)
	jsonAsBytes, _ := json.Marshal(certIndex)
	err = stub.PutState(certIndexStr, jsonAsBytes)						//store name of cert

	fmt.Println("- end init cert")
	return nil, nil
}

// ============================================================================================================================
// Set User Permission on certificate
// ============================================================================================================================
func (t *CertificateChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       1
	// "name", "renju"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	certAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Cert{}
	json.Unmarshal(certAsBytes, &res)										//un stringify it aka JSON.parse()
	res.User = args[1]														//change the user

	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the cert with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end set user")
	return nil, nil
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}
