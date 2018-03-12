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
MIICHTCCAcOgAwIBAgIULxGE/2ikZQrtNrAf4mPLizDUIlcwCgYIKoZIzj0EAwIw
czELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMT
E2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTgwMzEyMDIzMTAwWhcNMTkwMzEyMDIz
MTAwWjARMQ8wDQYDVQQDEwZob3R0YWIwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AAQ/NZxhlJrCp1GyXz4q/PeACnPyqHXyavTzObxT82M69LqBB4deF1R8AzDGl4nt
QI0SvFAl8NnvpYOcI4Oi1FKpo4GWMIGTMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMB
Af8EAjAAMB0GA1UdDgQWBBQh6Fn8/2aYygU0LGXIb1d+ywNdtzArBgNVHSMEJDAi
gCBCOaoNzXba7ri6DNpwGFHRRQTTGq0bLd3brGpXNl5JfDAnBggqAwQFBgcIAQQb
eyJhdHRycyI6eyJyb2xlIjoibWVtYmVyIn19MAoGCCqGSM49BAMCA0gAMEUCIQCi
rRgvug+UrtVtpxVAdDypV2vKIrvgOudyJL+/t85JDQIgL9r6E34ARd535Lc/W+7Y
gz9ZYCeoEANEc7cTvE2v+5s=
-----END CERTIFICATE-----
`

// testing.TB is an interface point to T and B, so it is pointer already
func createStub(t testing.TB, name string) *MockStub {
  scc := new(SimpleChaincode)
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

  Mspid, _ := cid.GetMSPID(stub)
  attrVal, found, err := cid.GetAttributeValue(stub, "role")
  t.Logf("Mspid: %v, attrVal: %v, found:%v \n", Mspid, attrVal, found)

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

  checkInvoke(t, stub, []string{"update_name", "123456789", "Tu Pham Thanh"})
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

// func Test_Buying(t *testing.T) {

// }

// func Test_TokenExchange(t *testing.T) {

// }
