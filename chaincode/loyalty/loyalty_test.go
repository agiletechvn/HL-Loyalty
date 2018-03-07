package main

import (
  "fmt"
  "github.com/golang/protobuf/proto"
  "github.com/hyperledger/fabric/core/chaincode/lib/cid"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/hyperledger/fabric/protos/msp"
  "testing"
)

const certWithAttrs = `-----BEGIN CERTIFICATE-----
MIIB6TCCAY+gAwIBAgIUHkmY6fRP0ANTvzaBwKCkMZZPUnUwCgYIKoZIzj0EAwIw
GzEZMBcGA1UEAxMQZmFicmljLWNhLXNlcnZlcjAeFw0xNzA5MDgwMzQyMDBaFw0x
ODA5MDgwMzQyMDBaMB4xHDAaBgNVBAMTE015VGVzdFVzZXJXaXRoQXR0cnMwWTAT
BgcqhkjOPQIBBggqhkjOPQMBBwNCAATmB1r3CdWvOOP3opB3DjJnW3CnN8q1ydiR
dzmuA6A2rXKzPIltHvYbbSqISZJubsy8gVL6GYgYXNdu69RzzFF5o4GtMIGqMA4G
A1UdDwEB/wQEAwICBDAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTYKLTAvJJK08OM
VGwIhjMQpo2DrjAfBgNVHSMEGDAWgBTEs/52DeLePPx1+65VhgTwu3/2ATAiBgNV
HREEGzAZghdBbmlscy1NYWNCb29rLVByby5sb2NhbDAmBggqAwQFBgcIAQQaeyJh
dHRycyI6eyJhdHRyMSI6InZhbDEifX0wCgYIKoZIzj0EAwIDSAAwRQIhAPuEqWUp
svTTvBqLR5JeQSctJuz3zaqGRqSs2iW+QB3FAiAIP0mGWKcgSGRMMBvaqaLytBYo
9v3hRt1r8j8vN0pMcg==
-----END CERTIFICATE-----
`

func checkInit(t *testing.T, stub *MockStub, args []string) {
  res := stub.MockInit("1", args)
  if res.Status != shim.OK {
    fmt.Println("Init failed", string(res.Message))
    t.FailNow()
  }
}

func checkInvoke(t *testing.T, stub *MockStub, name string, value string) {
  res := stub.MockInvoke("1", []string{name})
  if res.Status != shim.OK {
    fmt.Println("Query", name, "failed", string(res.Message))
    t.FailNow()
  }

  if res.Payload == nil {
    fmt.Println("Query", name, "failed to get value")
    t.FailNow()
  }

  if string(res.Payload) != value {
    fmt.Println("Query value", name, "was not", value, "as expected")
    t.FailNow()
  }
}

func Test_Cert_Attrs(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := NewMockStub("loyalty", scc)

  sid := &msp.SerializedIdentity{
    Mspid:   "Org1MSP",
    IdBytes: []byte(certWithAttrs),
  }
  b, err := proto.Marshal(sid)
  if err != nil {
    t.Fatalf("Cannot create msp: %v", err)
  }
  stub.creator = b

  sinfo, _ := cid.New(stub)

  attrVal, found, err := sinfo.GetAttributeValue("attr1")
  fmt.Printf("attrVal: %v, err: %v, found:%b \n", attrVal, err, found)

  if err != nil {
    t.Errorf("Error getting Unique ID of the submitter of the transaction")
  } else if !found {
    t.Errorf("Attribute 'role' should be found in the submitter cert")
  } else if attrVal != "val1" {
    t.Errorf("Assert have failed; value was %v, not val1", attrVal)
  }
}

func Test_Init(t *testing.T) {

  scc := new(SimpleChaincode)
  stub := NewMockStub("loyalty", scc)
  checkInit(t, stub, []string{"HT1234567", "Ha Noi", "HT1234568", "Hai Phong"})
}

func Test_Invoke(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := NewMockStub("loyalty", scc)
  checkInit(t, stub, []string{})

  checkInvoke(t, stub, "ping", "pong")
}
