package main

import (
  "fmt"
  "github.com/golang/protobuf/proto"
  "github.com/hyperledger/fabric/core/chaincode/lib/cid"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/hyperledger/fabric/protos/msp"
  "strings"
  "testing"
)

const certWithAttrs = `-----BEGIN CERTIFICATE-----
MIICGzCCAcKgAwIBAgIUduKmIyLqX4u0RQG6RGg0pqJv8kcwCgYIKoZIzj0EAwIw
czELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMT
E2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTgwMzE0MDgwMzAwWhcNMTkwMzE0MDgw
MzAwWjAQMQ4wDAYDVQQDEwV1c2VyMTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BOm13HxAE5ro/hnmYRz2em0zvPRxqRs6OLpl7OduG6IwrF4SoMsj5q02c+/B/mhG
FqiFxcHtHQeZzD3MUVK9YMejgZYwgZMwDgYDVR0PAQH/BAQDAgeAMAwGA1UdEwEB
/wQCMAAwHQYDVR0OBBYEFK/eK+/uCNlioUh1K6gMQfa/vswhMCsGA1UdIwQkMCKA
IEI5qg3NdtruuLoM2nAYUdFFBNMarRst3dusalc2Xkl8MCcGCCoDBAUGBwgBBBt7
ImF0dHJzIjp7InJvbGUiOiJtZW1iZXIifX0wCgYIKoZIzj0EAwIDRwAwRAIgG4r4
F4iemSkt7IxdrNhcJMnjPo4KJfxyFjhCdEvBQZICIAuTdCnxHg0WFj2Pr5pj6R7x
P06pc88pFqMnNdYJHk1U
-----END CERTIFICATE-----
`

// testing.TB is an interface point to T and B, so it is pointer already
func createStub(t testing.TB, name string) *MockStub {
  // new equal &
  scc := &SimpleChaincode{
    CashbackDecimal: 100,
    TokenDecimal:    1000000000,
    TokenSymbol:     "HTN",
    TokenName:       "Hottab Token",
    TotalSupply:     100000000,
  }

  stub := NewMockStub(name, scc)

  sid := &msp.SerializedIdentity{
    Mspid:   "Org1MSP",
    IdBytes: []byte(certWithAttrs),
  }
  b, err := proto.Marshal(sid)
  if err != nil {
    t.Fatalf("Cannot create msp: %v", err)
  }
  // by default hyperledger using protobuf
  stub.creator = b
  return stub
}

func checkInit(t *testing.T, stub *MockStub, args []string) {
  res := stub.MockInit("1", args)
  if res.Status != shim.OK {
    t.Errorf("Init failed: %s", res.Message)
    t.FailNow()
  }
}

func checkInvoke(t *testing.T, stub *MockStub, args []string, params ...string) {
  res := stub.MockInvoke("1", args)
  if res.Status != shim.OK {
    t.Errorf("Query %v failed, message: %s", args, res.Message)
    t.FailNow()
  }

  if value := strings.Join(params, ""); string(res.Payload) != value {
    t.Errorf("Query value %v got\n %s \nwas not\n %s as expected", args, res.Payload, value)
    t.FailNow()
  }
}

func Test_Cert_Attrs(t *testing.T) {

  stub := createStub(t, "loyalty")
  t.Logf("SmartContract: %+v", stub.GetChaincode())
  Mspid, _ := cid.GetMSPID(stub)
  cert, _ := cid.GetX509Certificate(stub)
  attrVal, found, err := cid.GetAttributeValue(stub, "role")
  t.Logf("Mspid: %v, commonName: %v, attrVal: %v, found:%v \n", Mspid, cert.Issuer.CommonName, attrVal, found)

  if err != nil {
    t.Errorf("Error getting Unique ID of the submitter of the transaction: %v", err)
  } else if !found {
    t.Errorf("Attribute 'role' should be found in the submitter cert")
  } else if attrVal != "member" {
    t.Errorf("Assert have failed; value was %v, not member", attrVal)
  }
}

func Test_Init(t *testing.T) {
  stub := createStub(t, "loyalty")
  checkInit(t, stub, []string{})
  checkInvoke(t, stub, []string{"ping"}, "pong")
}

func Test_Invoke(t *testing.T) {
  stub := createStub(t, "loyalty")
  checkInit(t, stub, []string{"123456789", "Ha Noi", "123456788", "Hai Phong"})
  // check get_pos_details function
  checkInvoke(t, stub, []string{"get_pos_details", "123456789"}, `{"posId":"123456789","posName":"Ha Noi","status":true,"percentage":5}`)
}

func Benchmark_Invoke(b *testing.B) {
  stub := createStub(b, "loyalty")
  stub.MockInit("1", []string{"123456789", "Ha Noi", "123456788", "Hai Phong"})

  for i := 0; i < b.N; i++ {
    stub.MockInvoke("1", []string{"get_pos_details", "123456789"})
  }
}

func Test_Items(t *testing.T) {
  stub := createStub(t, "loyalty")
  checkInit(t, stub, []string{})

  checkInvoke(t, stub, []string{"create_item", "123456789"})
  checkInvoke(t, stub, []string{"get_item_details", "123456789"}, `{"itemId":"123456789","posId":"0","itemName":"UNDEFINED","price":0}`)
}

func Benchmark_Items(b *testing.B) {
  stub := createStub(b, "loyalty")
  stub.MockInit("1", []string{})

  for i := 1; i <= b.N; i++ {
    id := fmt.Sprintf("%09d", i)
    name := fmt.Sprintf("Item %d", i)
    stub.MockInvoke("1", []string{"create_item", id, name})
  }

}

func Test_Customers(t *testing.T) {
  stub := createStub(t, "loyalty")
  checkInit(t, stub, []string{})

  checkInvoke(t, stub, []string{"create_customer", "123456789"})
  checkInvoke(t, stub, []string{"get_customer_details", "123456789"}, `{"customerID":"123456789","name":"`+CUSTOMER_PREFIX+`123456789","address":"UNDEFINED","cashback":0,"token":0,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`)

  checkInvoke(t, stub, []string{"update_customer_name", "123456789", "Tu Pham Thanh"})
  checkInvoke(t, stub, []string{"get_customer_details", "123456789"}, `{"customerID":"123456789","name":"Tu Pham Thanh","address":"UNDEFINED","cashback":0,"token":0,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`)

  checkInvoke(t, stub, []string{"create_customer", "123456788"})
  checkInvoke(t, stub, []string{"create_customer", "123456888"})
  // get range between startKey and endKey
  checkInvoke(t, stub, []string{"get_customers", "123456788", "123456789"},
    "[",
    `{"customerID":"123456788","name":"`+CUSTOMER_PREFIX+`123456788","address":"UNDEFINED","cashback":0,"token":0,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`,
    ",",
    `{"customerID":"123456789","name":"Tu Pham Thanh","address":"UNDEFINED","cashback":0,"token":0,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`,
    "]",
  )

}

func Benchmark_CreateCustomer(b *testing.B) {
  stub := createStub(b, "loyalty")
  stub.MockInit("1", []string{})

  for i := 1; i <= b.N; i++ {
    id := fmt.Sprintf("%09d", i)
    stub.MockInvoke("1", []string{"create_customer", id})
  }
}

func Test_Buying(t *testing.T) {
  stub := createStub(t, "loyalty")
  checkInit(t, stub, []string{"123456789", "Ha Noi"})
  checkInvoke(t, stub, []string{"update_percentage", "123456789", "10"})

  checkInvoke(t, stub, []string{"create_customer", "123456789"})
  checkInvoke(t, stub, []string{"update_customer_name", "123456789", "Tu Pham Thanh"})

  // price is $ multi by factor 100 decimal, so that we do not have to deal with float number
  checkInvoke(t, stub, []string{"create_item", "123456789"})
  checkInvoke(t, stub, []string{"update_item_name", "123456789", "Donut"})
  checkInvoke(t, stub, []string{"update_pos_id", "123456789", "123456789"})
  checkInvoke(t, stub, []string{"update_price", "123456789", "500"})
  checkInvoke(t, stub, []string{"get_item_details", "123456789"}, `{"itemId":"123456789","posId":"123456789","itemName":"Donut","price":500}`)

  // buy something, x10 product get 1 more with loyalty percentage 10%,
  for i := 0; i < 11; i++ {
    checkInvoke(t, stub, []string{"buy_item_by_money", "123456789", "123456789"})
  }
  checkInvoke(t, stub, []string{"get_customer_details", "123456789"}, `{"customerID":"123456789","name":"Tu Pham Thanh","address":"UNDEFINED","cashback":550,"token":0,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`)
  checkInvoke(t, stub, []string{"buy_item_by_wallet", "123456789", "123456789"})
  checkInvoke(t, stub, []string{"get_customer_details", "123456789"}, `{"customerID":"123456789","name":"Tu Pham Thanh","address":"UNDEFINED","cashback":50,"token":0,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`)
}

func Test_Transaction(t *testing.T) {
  stub := createStub(t, "loyalty")
  checkInit(t, stub, []string{"123456789", "Ha Noi"})
  checkInvoke(t, stub, []string{"update_percentage", "123456789", "10"})

  checkInvoke(t, stub, []string{"create_customer", "123456789"})
  checkInvoke(t, stub, []string{"update_customer_name", "123456789", "Tu Pham Thanh"})

  checkInvoke(t, stub, []string{"create_item", "123456789"})
  checkInvoke(t, stub, []string{"update_item_name", "123456789", "Donut"})
  checkInvoke(t, stub, []string{"update_pos_id", "123456789", "123456789"})
  checkInvoke(t, stub, []string{"update_price", "123456789", "100"})
  checkInvoke(t, stub, []string{"get_item_details", "123456789"}, `{"itemId":"123456789","posId":"123456789","itemName":"Donut","price":100}`)

  checkInvoke(t, stub, []string{"buy_item_by_money", "123456789", "123456789"})
  checkInvoke(t, stub, []string{"reward_cashback", "123456789", "150"})
  checkInvoke(t, stub, []string{"buy_item_by_wallet", "123456789", "123456789"})
  checkInvoke(t, stub, []string{"get_customer_details", "123456789"}, `{"customerID":"123456789","name":"Tu Pham Thanh","address":"UNDEFINED","cashback":60,"token":0,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`)

  // token method
  checkInvoke(t, stub, []string{"reward_token", "123456789", "2000"})
  checkInvoke(t, stub, []string{"get_customer_details", "123456789"}, `{"customerID":"123456789","name":"Tu Pham Thanh","address":"UNDEFINED","cashback":60,"token":2000,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`)
  checkInvoke(t, stub, []string{"burn_token", "123456789", "1000"})
  checkInvoke(t, stub, []string{"get_customer_details", "123456789"}, `{"customerID":"123456789","name":"Tu Pham Thanh","address":"UNDEFINED","cashback":60,"token":1000,"email":"UNDEFINED","phone":"UNDEFINED","status":true}`)

  checkInvoke(t, stub, []string{"get_market_info"}, `{"CashbackDecimal":100, "TokenDecimal":1000000000, "TokenSymbol":"HTN", "TokenName":"Hottab Token", "TotalSupply":100000000, "CirculatingSupply":1000}`)

}

// func Test_TokenExchange(t *testing.T) {

// }
