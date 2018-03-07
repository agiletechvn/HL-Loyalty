/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
  "encoding/json"
  "errors"
  "fmt"
  "github.com/hyperledger/fabric/core/chaincode/lib/cid"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/hyperledger/fabric/protos/peer"
  "regexp"
  "strings"
)

var logger = shim.NewLogger("LoyaltyChaincode")

//==============================================================================================================================
//   Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//             user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const AUTHORITY = "regulator"
const HOTEL = "hotel"
const AIRLINES = "airlines"
const CUSTOMER = "customer"
const RESTAURANT = "restaurant"
const VENDOR = "vendor"

//==============================================================================================================================
//   Structure Definitions
//==============================================================================================================================
//  Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//        and other HyperLedger functions)
//==============================================================================================================================
type SimpleChaincode struct {
}

//==============================================================================================================================
//  Customer - Defines the structure for a customer object. JSON on right tells it what JSON fields to map to
//        that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Customer struct {
  CustomerID string `json:"customerID"`
  Name       string `json:"name"`
  Address    string `json:"address"`
  Cashback   int    `json:"cashback"`
  Email      string `json:"email"`
  Phone      string `json:"phone"`
  Status     bool   `json:"status"`
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
//  CustomerID Holder - Defines the structure that holds all the customerIDs for Customer that have been created.
//        Used as an index when querying all vehicles.
//==============================================================================================================================

type CustomerID_Holder struct {
  Customers []string `json:"customers"`
}

//==============================================================================================================================
//  POS ID Holder - Defines the structure that holds all the POS IDs for Customer that have been created.
//        Used as an index when querying all vehicles.
//==============================================================================================================================

type PoSID_Holder struct {
  PoSIDs []string `json:"posIDs"`
}

//==============================================================================================================================
//  Item ID Holder - Defines the structure that holds all the Item IDs for Customer that have been created.
//        Used as an index when querying all vehicles.
//==============================================================================================================================

type ItemID_Holder struct {
  ItemIDs []string `json:"itemIDIDs"`
}

//==============================================================================================================================
//  Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
  // by default it is init function so no need to pass
  args := stub.GetStringArgs()
  // var customerIDs CustomerID_Holder

  // bytes, err := json.Marshal(customerIDs)

  // if err != nil {
  //   return nil, errors.New("Error creating pos record")
  // }

  // err = stub.PutState("customerIDs", bytes)

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
//   General Functions
//==============================================================================================================================
//   get_ecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//         for that user. Returns the ecert as retrieved including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) get_ecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

  ecert, err := stub.GetState(name)

  if err != nil {
    return nil, errors.New("Couldn't retrieve ecert for user " + name)
  }

  return ecert, nil
}

//==============================================================================================================================
//   read_cert_attribute - Retrieves the attribute name of the certificate.
//          Returns the attribute as a string.
//==============================================================================================================================

func (t *SimpleChaincode) read_cert_attribute(stub shim.ChaincodeStubInterface, name string) (string, error) {
  val, ok, err := cid.GetAttributeValue(stub, "attr1")
  if err != nil {
    return "", err
  }
  if !ok {
    return "", errors.New("The attribute is not found")
  }
  return val, nil
}

//==============================================================================================================================
//   get_username - Retrieves the username of the user who invoked the chaincode.
//          Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

  username, err := t.read_cert_attribute(stub, "username")
  if err != nil {
    return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error())
  }
  return username, nil
}

//==============================================================================================================================
//   check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
//              certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub shim.ChaincodeStubInterface) (string, error) {
  affiliation, err := t.read_cert_attribute(stub, "role")
  if err != nil {
    return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error())
  }
  return affiliation, nil

}

//==============================================================================================================================
//   get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//           name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error) {

  user, err := t.get_username(stub)

  // if err != nil { return "", "", err }

  // ecert, err := t.get_ecert(stub, user);

  // if err != nil { return "", "", err }

  affiliation, err := t.check_affiliation(stub)

  if err != nil {
    return "", "", err
  }

  return user, affiliation, nil
}

//==============================================================================================================================
//   retrieve_customer - Gets the state of the data at customerID in the ledger then converts it from the stored
//          JSON into the Customer struct for use in the contract. Returns the Vehcile struct.
//          Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_customer(stub shim.ChaincodeStubInterface, customerID string) (Customer, error) {

  var v Customer
  bytes, err := stub.GetState(customerID)

  if err != nil {
    fmt.Printf("RETRIEVE_CUSTOMER: Failed to invoke Customer_code: %s", err)
    return v, errors.New("RETRIEVE_CUSTOMER: Error retrieving Customer with customerID = " + customerID)
  }

  err = json.Unmarshal(bytes, &v)

  if err != nil {
    fmt.Printf("RETRIEVE_CUSTOMER: Corrupt Customer record "+string(bytes)+": %s", err)
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

  bytes, err := stub.GetState(itemID)

  if err != nil {
    fmt.Printf("RETRIEVE_ITEM: Failed to invoke ItemID: %s", err)
    return v, errors.New("RETRIEVE_ITEM: Error retrieving Item with ItemID = " + itemID)
  }

  err = json.Unmarshal(bytes, &v)

  if err != nil {
    fmt.Printf("RETRIEVE_ITEM: Corrupt Item record "+string(bytes)+": %s", err)
    return v, errors.New("RETRIEVE_ITEM: Corrupt Item record" + string(bytes))
  }

  return v, nil
}

//==============================================================================================================================
//   retrieve_pos - Gets the state of the data at posID in the ledger then converts it from the stored
//          JSON into the Item struct for use in the contract. Returns the Vehcile struct.
//          Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_pos(stub shim.ChaincodeStubInterface, posID string) (PoS, error) {

  var v PoS

  bytes, err := stub.GetState(posID)

  if err != nil {
    fmt.Printf("RETRIEVE_PoS: Failed to invoke posID: %s", err)
    return v, errors.New("RETRIEVE_ITEM: Error retrieving PoS with posID = " + posID)
  }

  err = json.Unmarshal(bytes, &v)

  if err != nil {
    fmt.Printf("RETRIEVE_PoS: Corrupt PoS record "+string(bytes)+": %s", err)
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
    fmt.Printf("SAVE_CHANGES: Error converting customer record: %s", err)
    return false, errors.New("Error converting customer record")
  }

  err = stub.PutState(v.CustomerID, bytes)

  if err != nil {
    fmt.Printf("SAVE_CHANGES: Error storing customer record: %s", err)
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
    fmt.Printf("SAVE_CHANGES: Error converting pos record: %s", err)
    return false, errors.New("Error converting pos record")
  }

  err = stub.PutState(v.PoSName, bytes)

  if err != nil {
    fmt.Printf("SAVE_CHANGES: Error storing pos record: %s", err)
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
    fmt.Printf("SAVE_CHANGES: Error converting pos record: %s", err)
    return false, errors.New("Error converting pos record")
  }

  err = stub.PutState(v.ItemName, bytes)

  if err != nil {
    fmt.Printf("SAVE_CHANGES: Error storing pos record: %s", err)
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
  return shim.Success([]byte(result))
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
    if len(args) != 1 {
      fmt.Printf("Incorrect number of arguments passed")
      return nil, errors.New("QUERY: Incorrect number of arguments passed")
    }
    v, err := t.retrieve_customer(stub, args[0])
    if err != nil {
      fmt.Printf("QUERY: Error retrieving v5c: %s", err)
      return nil, errors.New("QUERY: Error retrieving v5c " + err.Error())
    }
    return t.get_customer_details(stub, v)

  case "check_unique_customer":
    return t.check_unique_customer(stub, args[0])

  case "get_customers":
    return t.get_customers(stub)

  case "ping":
    return t.ping(stub)

  case "create_customer":
    return t.create_customer(stub, args[0])

  default:
    argPos := 0
    v, err := t.retrieve_customer(stub, args[argPos])

    if err != nil {
      fmt.Printf("INVOKE: Error retrieving Customer: %s", err)
      return nil, errors.New("Error retrieving customer")
    }
    if strings.Contains(function, "update") == false && function != "delete_customer" {
      if function == "buy_item_by_money" {
        argPos := 2
        i, err := t.retrieve_item(stub, args[argPos])
        if err != nil {
          fmt.Printf("INVOKE: Error retrieving Item: %s", err)
          return nil, errors.New("Error retrieving Item")
        }
        return t.buy_item_by_money(stub, v, i)
      } else if function == "buy_item_by_wallet" {
        argPos := 2
        i, err := t.retrieve_item(stub, args[argPos])
        if err != nil {
          fmt.Printf("INVOKE: Error retrieving Item: %s", err)
          return nil, errors.New("Error retrieving Item")
        }
        return t.buy_item_by_wallet(stub, v, i)
      }

    } else if function == "update_name" {
      return t.update_name(stub, v, args[0])
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
  v := Customer{
    CustomerID: customerID,
    Name:       customerID,
    Address:    "UNDEFINED",
    Cashback:   0,
    Email:      "UNDEFINED",
    Phone:      "UNDEFINED",
    Status:     true,
  }

  matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(customerID)) // matched = true if the customerId passed fits format of two letters followed by seven digits
  if err != nil {
    fmt.Printf("CREATE_CUSTOMER: Invalid customerID: %s", err)
    return nil, errors.New("Invalid customerID")
  }

  if customerID == "" || matched == false {
    fmt.Printf("CREATE_CUSTOMER: Invalid customerID provided")
    return nil, errors.New("Invalid customerID provided " + customerID)
  }

  record, err := stub.GetState(v.CustomerID) // If not an error then a record exists so cant create a new car with this customerId as it must be unique
  if record != nil {
    return nil, errors.New("Customer already exists")
  }

  _, err = t.save_changes(stub, v)
  if err != nil {
    fmt.Printf("CREATE_CUSTOMER: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  bytes, err := stub.GetState("customerIDs")
  if err != nil {
    return nil, errors.New("Unable to get customerID")
  }
  var customerIDs CustomerID_Holder
  err = json.Unmarshal(bytes, &customerIDs)
  if err != nil {
    return nil, errors.New("Corrupt Customer record")
  }
  customerIDs.Customers = append(customerIDs.Customers, customerID)
  bytes, err = json.Marshal(customerIDs)
  if err != nil {
    fmt.Print("Error creating Customer record")
  }
  err = stub.PutState("customerIDs", bytes)
  if err != nil {
    return nil, errors.New("Unable to put the state")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_name
//=================================================================================================================================
func (t *SimpleChaincode) update_name(stub shim.ChaincodeStubInterface, v Customer, new_value string) ([]byte, error) {

  if v.Status == true {
    v.Name = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_NAME: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_address
//=================================================================================================================================
func (t *SimpleChaincode) update_address(stub shim.ChaincodeStubInterface, v Customer, new_value string) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  if v.Status == true {
    v.Address = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_ADDRESS: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_cashback
//=================================================================================================================================
func (t *SimpleChaincode) update_cashback(stub shim.ChaincodeStubInterface, v Customer, new_value int) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  if v.Status == true {
    v.Cashback = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_CASHBACK: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_email
//=================================================================================================================================
func (t *SimpleChaincode) update_email(stub shim.ChaincodeStubInterface, v Customer, new_value string) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  if v.Status == true {
    v.Email = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_EMAIL: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   Create PoS - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_pos(stub shim.ChaincodeStubInterface, posID string, posName string) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  v := PoS{
    PoSID:             posID,
    PoSName:           posName,
    Status:            true,
    LoyaltyPercentage: 5,
  }

  matched, err := regexp.Match("^[A-z]{2}[0-9]{7}", []byte(posID)) // matched = true if the customerId passed fits format of two letters followed by seven digits
  if err != nil {
    fmt.Printf("CREATE_POS: Invalid posID: %s", err)
    return nil, errors.New("Invalid posID")
  }

  if posID == "" || matched == false {
    fmt.Printf("CREATE_POS: Invalid posID provided")
    return nil, errors.New("Invalid posID provided")
  }

  record, err := stub.GetState(v.PoSID) // If not an error then a record exists so cant create a new car with this CustomerID as it must be unique
  if record != nil {
    return nil, errors.New("POS already exists")
  }

  _, err = t.save_changes_pos(stub, v)
  if err != nil {
    fmt.Printf("CREATE_POS: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  bytes, err := stub.GetState("posIDs")
  if err != nil {
    return nil, errors.New("Unable to get PoSID")
  }
  var posIDs PoSID_Holder
  if len(bytes) > 0 {
    err = json.Unmarshal(bytes, &posIDs)
    if err != nil {
      return nil, errors.New("Corrupt PoS record")
    }
  } else {
    // by default init an empty posIDs
    posIDs = PoSID_Holder{}
  }
  // then append the item
  posIDs.PoSIDs = append(posIDs.PoSIDs, posID)
  bytes, err = json.Marshal(posIDs)
  if err != nil {
    fmt.Print("Error creating PoS record")
  }
  err = stub.PutState("posIDs", bytes)
  if err != nil {
    return nil, errors.New("Unable to put the state")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_posname
//=================================================================================================================================
func (t *SimpleChaincode) update_posname(stub shim.ChaincodeStubInterface, v PoS, new_value string) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  if v.Status == true {
    v.PoSName = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes_pos(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_POSNAME: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_percentage
//=================================================================================================================================
func (t *SimpleChaincode) update_percentage(stub shim.ChaincodeStubInterface, v PoS, new_value int) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  if v.Status == true {
    v.LoyaltyPercentage = new_value
  } else {
    return nil, errors.New(fmt.Sprint("Not found"))
  }

  _, err := t.save_changes_pos(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_PERCENTAGE: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   Create PoS - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_item(stub shim.ChaincodeStubInterface, itemID string) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  v := Item{
    ItemID:   itemID,
    PoSID:    "0",
    ItemName: "UNDEFINED",
    Price:    500,
  }

  matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(itemID)) // matched = true if the customerId passed fits format of two letters followed by seven digits
  if err != nil {
    fmt.Printf("CREATE_ITEM: Invalid itemID: %s", err)
    return nil, errors.New("Invalid itemID")
  }

  if itemID == "" || matched == false {
    fmt.Printf("CREATE_ITEM: Invalid itemID provided")
    return nil, errors.New("Invalid itemID provided")
  }

  record, err := stub.GetState(v.ItemID) // If not an error then a record exists so cant create a new car with this CustomerID as it must be unique
  if record != nil {
    return nil, errors.New("Item already exists")
  }

  _, err = t.save_changes_item(stub, v)
  if err != nil {
    fmt.Printf("CREATE_POS: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  bytes, err := stub.GetState("itemID")
  if err != nil {
    return nil, errors.New("Unable to get ItemID")
  }
  var itemIDs ItemID_Holder
  err = json.Unmarshal(bytes, &itemIDs)
  if err != nil {
    return nil, errors.New("Corrupt Item record")
  }
  itemIDs.ItemIDs = append(itemIDs.ItemIDs, itemID)
  bytes, err = json.Marshal(itemIDs)
  if err != nil {
    fmt.Print("Error creating Item record")
  }
  err = stub.PutState("itemIDs", bytes)
  if err != nil {
    return nil, errors.New("Unable to put the state")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_item_name
//=================================================================================================================================
func (t *SimpleChaincode) update_item_name(stub shim.ChaincodeStubInterface, v Item, new_value string) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  v.ItemName = new_value

  _, err := t.save_changes_item(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_ITEM_NAME: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_posid
//=================================================================================================================================
func (t *SimpleChaincode) update_posid(stub shim.ChaincodeStubInterface, v Item, new_value string) ([]byte, error) {

  // caller, caller_affiliation, err :=  t.get_caller_data(stub)

  v.PoSID = new_value

  _, err := t.save_changes_item(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_POSID: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   update_price
//=================================================================================================================================
func (t *SimpleChaincode) update_price(stub shim.ChaincodeStubInterface, v Item, new_value int) ([]byte, error) {

  // caller string, caller_affiliation string,

  v.Price = new_value

  _, err := t.save_changes_item(stub, v)
  if err != nil {
    fmt.Printf("UPDATE_PRICE: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil
}

//=================================================================================================================================
//   Read Functions
//=================================================================================================================================
//   get_customer_details
//=================================================================================================================================
func (t *SimpleChaincode) get_customer_details(stub shim.ChaincodeStubInterface, v Customer) ([]byte, error) {

  bytes, err := json.Marshal(v)
  if err != nil {
    return nil, errors.New("GET_CUSTOMER_DETAILS: Invalid Customer object")
  }
  return bytes, nil
}

//=================================================================================================================================
//   get_customers
//=================================================================================================================================

func (t *SimpleChaincode) get_customers(stub shim.ChaincodeStubInterface) ([]byte, error) {
  bytes, err := stub.GetState("customerIDs")
  if err != nil {
    return nil, errors.New("Unable to get customerIDs")
  }
  var customerIDs CustomerID_Holder
  err = json.Unmarshal(bytes, &customerIDs)
  if err != nil {
    return nil, errors.New("Corrupt CustomerID_Holder")
  }
  result := "["
  var temp []byte
  var v Customer

  for _, customer := range customerIDs.Customers {

    v, err = t.retrieve_customer(stub, customer)

    if err != nil {
      return nil, errors.New("Failed to retrieve Customer")
    }

    temp, err = t.get_customer_details(stub, v)

    if err == nil {
      result += string(temp) + ","
    }
  }

  if len(result) == 1 {
    result = "[]"
  } else {
    result = result[:len(result)-1] + "]"
  }

  return []byte(result), nil
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
//   buy_item_by_money
//=================================================================================================================================
func (t *SimpleChaincode) buy_item_by_money(stub shim.ChaincodeStubInterface, v Customer, i Item) ([]byte, error) {

  if v.Status == true {
    p, err := t.retrieve_pos(stub, i.PoSID)
    if err != nil {
      fmt.Printf("INVOKE: Error retrieving PoS: %s", err)
      return nil, errors.New("Error retrieving PoS")
    }
    v.Cashback = v.Cashback + (p.LoyaltyPercentage*i.Price)/100
  } else { // Otherwise if there is an error
    fmt.Printf("buy_item_by_money: Customer Not Active")
    return nil, errors.New(fmt.Sprintf(" Customer Not Active."))
  }

  _, err := t.save_changes(stub, v) // Write new state
  if err != nil {
    fmt.Printf("buy_item_by_money: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil // We are Done
}

func (t *SimpleChaincode) buy_item_by_wallet(stub shim.ChaincodeStubInterface, v Customer, i Item) ([]byte, error) {

  if v.Status == true {
    if v.Cashback > i.Price {
      v.Cashback = v.Cashback - i.Price
    } else {
      fmt.Printf("buy_item_by_wallet: Not enough balance")
      return nil, errors.New(fmt.Sprintf(" Not enough balance."))
    }
  } else { // Otherwise if there is an error
    fmt.Printf("buy_item_by_wallet: Customer Not Active")
    return nil, errors.New(fmt.Sprintf(" Customer Not Active."))
  }
  _, err := t.save_changes(stub, v) // Write new state
  if err != nil {
    fmt.Printf("buy_item_by_wallet: Error saving changes: %s", err)
    return nil, errors.New("Error saving changes")
  }
  return nil, nil // We are Done
}

//=================================================================================================================================
//   Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

  err := shim.Start(new(SimpleChaincode))
  if err != nil {
    fmt.Printf("Error starting Chaincode: %s", err)
  }
}
