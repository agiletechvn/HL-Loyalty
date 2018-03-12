/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "github.com/hyperledger/fabric/core/chaincode/lib/cid"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/hyperledger/fabric/protos/peer"
  "regexp"
  "strconv"
  // "strings"
)

// using logger so that logs only appear if the ChaincodeLogger LoggingLevel is set to
// LogInfo or LogDebug
var logger = shim.NewLogger("LoyaltyChaincode")
var idPattern = regexp.MustCompile("^[0-9]{9}$")

const CUSTOMER_PREFIX = "CTM"
const POS_PREFIX = "POS"
const ITEM_PREFIX = "ITM"

//==============================================================================================================================
//   Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//             user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const AUTHORITY = "regulator"
const HOTEL = "hotel"
const AIRLINES = "airlines"
const RESTAURANT = "restaurant"
const VENDOR = "vendor"

//==============================================================================================================================
//   Structure Definitions
//==============================================================================================================================
//  Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//        and other HyperLedger functions)
//==============================================================================================================================
type SimpleChaincode struct {
  CashbackDecimal int
  TokenDecimal    int
  TokenSymbol     string
  TokenName       string
  TotalSupply     int64
}

//==============================================================================================================================
//  Customer - Defines the structure for a customer object. JSON on right tells it what JSON fields to map to
//        that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Customer struct {
  CustomerID string `json:"customerID"`
  Name       string `json:"name"`
  Address    string `json:"address"`
  // cashback received directly from buying items
  // token received from doing comment, rating
  Cashback int `json:"cashback"`
  // transfer back token to get cashback or stellar
  // loyalty point is token so that it incentivizes users directly
  // this is not ICO, so token is unlimited, transfer back token is like burning,
  // but we might allow transfer between 2 customers
  Token  int    `json:"token"`
  Email  string `json:"email"`
  Phone  string `json:"phone"`
  Status bool   `json:"status"`
}

//==============================================================================================================================
//  Point of Sales - Defines the structure that holds all the PoS for that have been created.
//        Used as an index when querying all vehicles.
//==============================================================================================================================

type PoS struct {
  PoSID             string `json:"posId"`
  PoSName           string `json:"posName"`
  Status            bool   `json:"status"`
  LoyaltyPercentage int    `json:"percentage"`
}

//==============================================================================================================================
//  Items - Items brought.
//==============================================================================================================================

type Item struct {
  ItemID   string `json:"itemId"`
  PoSID    string `json:"posId"`
  ItemName string `json:"itemName"`
  Price    int    `json:"price"`
}

//==============================================================================================================================
//  Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
  // by default it is init function so no need to pass
  args := stub.GetStringArgs()

  // add initial pos
  for i := 0; i < len(args); i = i + 2 {
    _, err := t.create_pos(stub, args[i], args[i+1])
    if err != nil {
      return shim.Error(err.Error())
    }
  }

  return shim.Success(nil)
}

//==============================================================================================================================
//   read_cert_attribute - Retrieves the attribute name of the certificate.
//          Returns the attribute as a string.
//==============================================================================================================================

func (t *SimpleChaincode) read_cert_attribute(stub shim.ChaincodeStubInterface, name string) (string, error) {
  val, ok, err := cid.GetAttributeValue(stub, name)

  if err != nil {
    return "", err
  }
  if !ok {
    return "", errors.New("The attribute is not found")
  }
  return val, nil
}

//==============================================================================================================================
//   check_role - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
//              certificates common name. The role is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_role(stub shim.ChaincodeStubInterface) (string, error) {
  role, err := t.read_cert_attribute(stub, "role")
  if err != nil {
    return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error())
  }
  return role, nil

}

//==============================================================================================================================
//   get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//           name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error) {

  mspID, _ := cid.GetMSPID(stub)

  role, err := t.check_role(stub)

  if err != nil {
    logger.Errorf("Couldn't get caller data, got error: %v", err)
    return "", "", err
  }

  logger.Infof("msp: %s, role: %s", mspID, role)

  return mspID, role, nil
}

//==============================================================================================================================
//   retrieve_customer - Gets the state of the data at customerID in the ledger then converts it from the stored
//          JSON into the Customer struct for use in the contract. Returns the Vehcile struct.
//          Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_customer(stub shim.ChaincodeStubInterface, customerID string) (Customer, error) {

  var v Customer
  bytes, err := t.get_customer_details(stub, customerID)

  if err != nil {
    return v, nil
  }

  err = json.Unmarshal(bytes, &v)

  if err != nil {
    logger.Infof("RETRIEVE_CUSTOMER: Corrupt Customer record: %s, got error: %s", bytes, err)
    return v, errors.New("RETRIEVE_CUSTOMER: Corrupt Customer record" + string(bytes))
  }

  return v, nil
}

//==============================================================================================================================
//   retrieve_item - Gets the state of the data at itemID in the ledger then converts it from the stored
//          JSON into the Item struct for use in the contract. Returns the Vehcile struct.
//          Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_item(stub shim.ChaincodeStubInterface, itemID string) (Item, error) {

  var v Item

  bytes, err := t.get_item_details(stub, itemID)

  if err != nil {
    return v, err
  }

  err = json.Unmarshal(bytes, &v)

  if err != nil {
    logger.Infof("RETRIEVE_ITEM: Corrupt Item record: %s, got error: %s", bytes, err)
    return v, errors.New("RETRIEVE_ITEM: Corrupt Item record" + string(bytes))
  }

  return v, nil
}

//==============================================================================================================================
//   get_pos_details - Gets the state of pos
//==============================================================================================================================
func (t *SimpleChaincode) get_pos_details(stub shim.ChaincodeStubInterface, posID string) ([]byte, error) {

  bytes, err := stub.GetState(POS_PREFIX + posID)
  if err != nil {
    logger.Infof("get_pos_details: Error retrieving pos: %s, with posID = %v", err, posID)
    return nil, errors.New("RETRIEVE_ITEM: Error retrieving PoS with posID = " + posID)
  }

  return bytes, nil
}

//==============================================================================================================================
//   get_item_details - Gets the state of item
//==============================================================================================================================
func (t *SimpleChaincode) get_item_details(stub shim.ChaincodeStubInterface, itemID string) ([]byte, error) {

  bytes, err := stub.GetState(ITEM_PREFIX + itemID)
  if err != nil {
    logger.Infof("get_item_details: Error retrieving pos: %s, with ItemID = %v", err, itemID)
    return nil, errors.New("RETRIEVE_ITEM: Error retrieving Item with ItemID = " + itemID)
  }
  return bytes, nil
}

//==============================================================================================================================
//   retrieve_pos - Gets the state of the data at posID in the ledger then converts it from the stored
//          JSON into the Item struct for use in the contract. Returns the Vehcile struct.
//          Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_pos(stub shim.ChaincodeStubInterface, posID string) (PoS, error) {

  var v PoS

  bytes, err := t.get_pos_details(stub, posID)

  if err != nil {
    return v, err
  }

  err = json.Unmarshal(bytes, &v)

  if err != nil {
    logger.Infof("RETRIEVE_PoS: Corrupt PoS record: %s, got error: %s", bytes, err)
    return v, errors.New("RETRIEVE_ITEM: Corrupt PoS record" + string(bytes))
  }

  return v, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Customer struct passed in a JSON format. Uses the shim file's
//          method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, v Customer) (bool, error) {

  bytes, err := json.Marshal(v)

  if err != nil {
    logger.Infof("SAVE_CHANGES: Error converting customer record: %s", err)
    return false, errors.New("Error converting customer record")
  }

  err = stub.PutState(CUSTOMER_PREFIX+v.CustomerID, bytes)

  if err != nil {
    logger.Infof("SAVE_CHANGES: Error storing customer record: %s", err)
    return false, errors.New("Error storing customer record")
  }

  return true, nil
}

//==============================================================================================================================
// save_changes_pos - Writes to the ledger the PoS struct passed in a JSON format. Uses the shim file's
//          method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes_pos(stub shim.ChaincodeStubInterface, v PoS) (bool, error) {

  bytes, err := json.Marshal(v)

  if err != nil {
    logger.Infof("SAVE_CHANGES: Error converting pos record: %s", err)
    return false, errors.New("Error converting pos record")
  }

  // update using posID as key
  err = stub.PutState(POS_PREFIX+v.PoSID, bytes)

  if err != nil {
    logger.Infof("SAVE_CHANGES: Error storing pos record: %s", err)
    return false, errors.New("Error storing pos record")
  }

  return true, nil
}

//==============================================================================================================================
// save_changes_item - Writes to the ledger the Item struct passed in a JSON format. Uses the shim file's
//          method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes_item(stub shim.ChaincodeStubInterface, v Item) (bool, error) {

  bytes, err := json.Marshal(v)

  if err != nil {
    logger.Infof("SAVE_CHANGES: Error converting pos record: %s", err)
    return false, errors.New("Error converting pos record")
  }

  err = stub.PutState(ITEM_PREFIX+v.ItemID, bytes)

  if err != nil {
    logger.Infof("SAVE_CHANGES: Error storing pos record: %s", err)
    return false, errors.New("Error storing pos record")
  }

  return true, nil
}

//==============================================================================================================================
//   Router Functions
//==============================================================================================================================
// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

  // Extract the function and args from the transaction proposal
  fn, args := stub.GetFunctionAndParameters()

  result, err := t.invoke(stub, fn, args)

  if err != nil {
    return shim.Error(err.Error())
  }

  // Return the result as success payload
  return shim.Success(result)
}

//==============================================================================================================================
//   Router Functions
//==============================================================================================================================
//  Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//      initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

  // go only runs the selected case, not all the cases that follow, switch case have the same indent
  switch function {

  case "get_customer_details":
    return t.get_customer_details(stub, args[0])

  case "get_pos_details":
    return t.get_pos_details(stub, args[0])

  case "get_item_details":
    return t.get_item_details(stub, args[0])

  case "check_unique_customer":
    return t.check_unique_customer(stub, args[0])

  case "get_customers":
    return t.get_customers(stub, args...)

  case "ping":
    return t.ping(stub)

  case "create_customer":
    return t.create_customer(stub, args[0])

  case "create_pos":
    return t.create_pos(stub, args[0], args[1])

  case "create_item":
    return t.create_item(stub, args[0])

  // other updates: pos, item
  case "update_item_name", "update_price", "update_pos_id":
    item, err := t.retrieve_item(stub, args[0])
    // check error first
    if err != nil {
      return nil, err
    }
    // finally do the rest
    if function == "update_price" {
      return t.update_price(stub, item, args[1])
    } else if function == "update_item_name" {
      return t.update_item_name(stub, item, args[1])
    } else {
      return t.update_pos_id(stub, item, args[1])
    }

  case "update_pos_name", "update_percentage":
    pos, err := t.retrieve_pos(stub, args[0])
    // check error first
    if err != nil {
      return nil, err
    }
    // do the rest
    if function == "update_pos_name" {
      return t.update_pos_name(stub, pos, args[0])
    } else {
      return t.update_percentage(stub, pos, args[1])
    }

  // process for existing customer
  default:
    v, err := t.retrieve_customer(stub, args[0])

    if err != nil {
      logger.Infof("INVOKE: Error retrieving Customer: %s", err)
      return nil, errors.New("Error retrieving customer")
    }

    // update customer name
    if function == "update_customer_name" {
      return t.update_customer_name(stub, v, args[1])
    }

    // buying item
    item, err := t.retrieve_item(stub, args[1])
    if err != nil {
      logger.Infof("INVOKE: Error retrieving Item: %s", err)
      return nil, errors.New("Error retrieving Item")
    }

    if function == "buy_item_by_money" {
      return t.buy_item_by_money(stub, v, item)
    } else if function == "buy_item_by_wallet" {
      return t.buy_item_by_wallet(stub, v, item)
    }

  }

  return nil, errors.New("Function of the name " + function + " doesn't exist.")
}

//=================================================================================================================================
//   Ping Function
//=================================================================================================================================
//   Pings the peer to keep the connection alive
//=================================================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
  return []byte("pong"), nil
}

//=================================================================================================================================
//   Create Function
//=================================================================================================================================
//   Create Customer - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_customer(stub shim.ChaincodeStubInterface, customerID string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  v := Customer{
    CustomerID: customerID,
    Name:       CUSTOMER_PREFIX + customerID,
    Address:    "UNDEFINED",
    Cashback:   0,
    Token:      0,
    Email:      "UNDEFINED",
    Phone:      "UNDEFINED",
    Status:     true,
  }

  // matched = true if the customerId passed fits format of two letters followed by seven digits
  matched := idPattern.Match([]byte(customerID))
  if customerID == "" || matched == false {
    logger.Errorf("CREATE_CUSTOMER: Invalid customerID provided: %v", customerID)
    return nil, errors.New("Invalid customerID provided " + customerID)
  }

  // If not an error then a record exists so cant create a new car with this customerId as it must be unique
  record, err := stub.GetState(CUSTOMER_PREFIX + v.CustomerID)
  if record != nil {
    logger.Errorf("Customer already exists %v", customerID)
    return nil, errors.New("Customer already exists")
  }

  _, err = t.save_changes(stub, v)
  if err != nil {
    logger.Errorf("CREATE_CUSTOMER: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }

  return nil, nil
}

//=================================================================================================================================
//   update_customer_name
//=================================================================================================================================
func (t *SimpleChaincode) update_customer_name(stub shim.ChaincodeStubInterface, v Customer, new_value string) ([]byte, error) {

  if v.Status == true {
    v.Name = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    logger.Infof("UPDATE_CUSTOMER_NAME: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_address
//=================================================================================================================================
func (t *SimpleChaincode) update_address(stub shim.ChaincodeStubInterface, v Customer, new_value string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  if v.Status == true {
    v.Address = new_value
  } else {
    return nil, errors.New("Not found")
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    logger.Infof("UPDATE_ADDRESS: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_cashback
//=================================================================================================================================
func (t *SimpleChaincode) update_cashback(stub shim.ChaincodeStubInterface, v Customer, new_value int) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  if v.Status == true {
    v.Cashback = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    logger.Infof("UPDATE_CASHBACK: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_email
//=================================================================================================================================
func (t *SimpleChaincode) update_email(stub shim.ChaincodeStubInterface, v Customer, new_value string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  if v.Status == true {
    v.Email = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    logger.Infof("UPDATE_EMAIL: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   Create PoS - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_pos(stub shim.ChaincodeStubInterface, posID string, posName string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  v := PoS{
    PoSID:             posID,
    PoSName:           posName,
    Status:            true,
    LoyaltyPercentage: 5,
  }

  // matched = true if the customerId passed fits format of two letters followed by seven digits
  matched := idPattern.Match([]byte(posID))

  if posID == "" || matched == false {
    logger.Infof("CREATE_POS: Invalid posID provided")
    return nil, errors.New("Invalid posID provided")
  }

  // If not an error then a record exists so cant create a new car with this CustomerID as it must be unique
  record, err := stub.GetState(POS_PREFIX + v.PoSID)
  if record != nil {
    return nil, errors.New("POS already exists")
  }

  _, err = t.save_changes_pos(stub, v)
  if err != nil {
    logger.Infof("CREATE_POS: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }

  return nil, nil
}

//=================================================================================================================================
//   update_pos_name
//=================================================================================================================================
func (t *SimpleChaincode) update_pos_name(stub shim.ChaincodeStubInterface, v PoS, new_value string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  if v.Status == true {
    v.PoSName = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes_pos(stub, v)
  if err != nil {
    logger.Infof("UPDATE_POS_NAME: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_percentage
//=================================================================================================================================
func (t *SimpleChaincode) update_percentage(stub shim.ChaincodeStubInterface, v PoS, percentage string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  new_value, err := strconv.Atoi(percentage)
  if err != nil {
    return nil, errors.New("Expecting integer value for item percentage")
  }

  if v.Status == true {
    v.LoyaltyPercentage = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err = t.save_changes_pos(stub, v)
  if err != nil {
    logger.Infof("UPDATE_PERCENTAGE: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   Create PoS - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_item(stub shim.ChaincodeStubInterface, itemID string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  // will update the code to set the price of an item
  v := Item{
    ItemID:   itemID,
    PoSID:    "0",
    ItemName: "UNDEFINED",
    Price:    0,
  }

  // matched = true if the customerId passed fits format of two letters followed by seven digits
  matched := idPattern.Match([]byte(itemID))

  if itemID == "" || matched == false {
    logger.Infof("CREATE_ITEM: Invalid itemID provided")
    return nil, errors.New("Invalid itemID provided")
  }

  // If not an error then a record exists so cant create a new car with this CustomerID as it must be unique
  record, err := stub.GetState(v.ItemID)
  if record != nil {
    return nil, errors.New("Item already exists")
  }

  _, err = t.save_changes_item(stub, v)
  if err != nil {
    logger.Infof("CREATE_ITEM: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }

  return nil, nil
}

//=================================================================================================================================
//   update_item_name
//=================================================================================================================================
func (t *SimpleChaincode) update_item_name(stub shim.ChaincodeStubInterface, v Item, new_value string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  v.ItemName = new_value

  _, err := t.save_changes_item(stub, v)
  if err != nil {
    logger.Infof("UPDATE_ITEM_NAME: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_pos_id
//=================================================================================================================================
func (t *SimpleChaincode) update_pos_id(stub shim.ChaincodeStubInterface, v Item, new_value string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  v.PoSID = new_value

  _, err := t.save_changes_item(stub, v)
  if err != nil {
    logger.Infof("UPDATE_POS_ID: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_price
//=================================================================================================================================
func (t *SimpleChaincode) update_price(stub shim.ChaincodeStubInterface, v Item, price string) ([]byte, error) {

  caller, role, _ := t.get_caller_data(stub)
  if role != "member" {
    return nil, errors.New(fmt.Sprintf("Unauthorized: %v", caller))
  }

  new_value, err := strconv.Atoi(price)
  if err != nil {
    return nil, errors.New("Expecting integer value for item price")
  }

  v.Price = new_value

  _, err = t.save_changes_item(stub, v)
  if err != nil {
    logger.Infof("UPDATE_PRICE: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   Read Functions
//=================================================================================================================================
//   get_customer_details
//=================================================================================================================================
func (t *SimpleChaincode) get_customer_details(stub shim.ChaincodeStubInterface, customerID string) ([]byte, error) {

  bytes, err := stub.GetState(CUSTOMER_PREFIX + customerID)
  if err != nil {
    logger.Infof("get_customer_details: Error retrieving customer: %s, with customerID = %v", err, customerID)
    return nil, errors.New("get_customer_details: Error retrieving Customer with customerID = " + customerID)
  }
  return bytes, nil
}

//=================================================================================================================================
//   get_customers
//=================================================================================================================================

func (t *SimpleChaincode) get_customers(stub shim.ChaincodeStubInterface, params ...string) ([]byte, error) {

  var startKey, endKey string
  if len(params) == 2 {
    startKey = CUSTOMER_PREFIX + params[0]
    endKey = CUSTOMER_PREFIX + params[1]
  } else if len(params) == 1 {
    startKey = CUSTOMER_PREFIX + "000000000"
    endKey = CUSTOMER_PREFIX + params[1]
  } else {
    startKey = CUSTOMER_PREFIX + "000000000"
    endKey = CUSTOMER_PREFIX + "999999999"
  }

  logger.Infof("StartKey: %v, EndKey: %v", startKey, endKey)

  resultsIterator, err := stub.GetStateByRange(startKey, endKey)
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  var buffer bytes.Buffer

  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    // Add a comma before array members, suppress it for the first array member
    if buffer.Len() == 0 {
      buffer.WriteString("[")
    } else {
      buffer.WriteString(",")
    }
    // Record is a JSON object in bytes, so we write as-is
    buffer.Write(queryResponse.Value)

  }
  buffer.WriteString("]")

  return buffer.Bytes(), nil
}

//=================================================================================================================================
//   check_unique_customer
//=================================================================================================================================
func (t *SimpleChaincode) check_unique_customer(stub shim.ChaincodeStubInterface, customerID string) ([]byte, error) {
  _, err := t.retrieve_customer(stub, customerID)
  if err == nil {
    return []byte("false"), errors.New("Customer is not unique")
  } else {
    return []byte("true"), nil
  }
}

//=================================================================================================================================
//   Transactions
//=================================================================================================================================

//=================================================================================================================================
//   buy_item_by_money: gain cashback by percentage of item
//=================================================================================================================================
func (t *SimpleChaincode) buy_item_by_money(stub shim.ChaincodeStubInterface, v Customer, i Item) ([]byte, error) {

  if v.Status == true {
    p, err := t.retrieve_pos(stub, i.PoSID)
    if err != nil {
      logger.Infof("INVOKE: Error retrieving PoS: %s", err)
      return nil, errors.New("Error retrieving PoS")
    }
    bonus := (p.LoyaltyPercentage * i.Price) / 100
    v.Cashback = v.Cashback + bonus
    logger.Infof("Customer %v received %d, balance is %d", v.CustomerID, bonus, v.Cashback)
  } else { // Otherwise if there is an error
    logger.Infof("buy_item_by_money: Customer Not Active")
    return nil, errors.New(fmt.Sprintf(" Customer Not Active."))
  }

  _, err := t.save_changes(stub, v) // Write new state
  if err != nil {
    logger.Infof("buy_item_by_money: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil // We are Done
}

//=================================================================================================================================
//   buy_item_by_wallet: using cashback to buy item, subtract the price of item
//   we will have others exchange function which convert loyaltypoints to token, stellar to token
//   and token can be converted to cashback, cashback should be like $ or something that can substract the item price
//=================================================================================================================================
func (t *SimpleChaincode) buy_item_by_wallet(stub shim.ChaincodeStubInterface, v Customer, i Item) ([]byte, error) {

  if v.Status == true {
    if v.Cashback > i.Price {
      v.Cashback = v.Cashback - i.Price
      logger.Infof("Customer %v spent %d, balance is %d", v.CustomerID, i.Price, v.Cashback)
    } else {
      logger.Infof("buy_item_by_wallet: Not enough balance")
      return nil, errors.New(fmt.Sprintf(" Not enough balance."))
    }
  } else { // Otherwise if there is an error
    logger.Infof("buy_item_by_wallet: Customer Not Active")
    return nil, errors.New(fmt.Sprintf("Customer Not Active."))
  }
  _, err := t.save_changes(stub, v) // Write new state
  if err != nil {
    logger.Infof("buy_item_by_wallet: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil // We are Done
}

//=================================================================================================================================
//   Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {
  err := shim.Start(&SimpleChaincode{
    CashbackDecimal: 100,
    TokenDecimal:    1000000000,
    TokenSymbol:     "HTN",
    TokenName:       "Hottab Token",
    TotalSupply:     100000000,
  })
  if err != nil {
    logger.Infof("Error starting Chaincode: %s", err)
  }
}
