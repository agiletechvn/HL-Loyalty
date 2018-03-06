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
MIIB8TCCAZigAwIBAgIUe3Y5111qvERcG7B6/2OQEvjIoZkwCgYIKoZIzj0EAwIw
czELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMT
E2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTgwMzA2MDQ1MjAwWhcNMTkwMzA2MDQ1
MjAwWjARMQ8wDQYDVQQDEwZob3R0YWIwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AAQdFyEgPvY7GqpGI0iotvCHOHx6dD+hPLhFxRb0lH2dOSp5P5EGyTr9tUdep2k+
/wX2gy5uY1q3Hvoz0WTk4I8Wo2wwajAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/
BAIwADAdBgNVHQ4EFgQUTy8NvjCrYhxOT5kMCW58U/ZjuiYwKwYDVR0jBCQwIoAg
QjmqDc122u64ugzacBhR0UUE0xqtGy3d26xqVzZeSXwwCgYIKoZIzj0EAwIDRwAw
RAIgHgoCu7O329PKwCzoQOJvkDIonhsHODHevg+WfKIv4EsCIAj2vA51ck2YQu0y
Zn3+DmFvbLBSFvm1Tgu01IFpuFoq
-----END CERTIFICATE-----
`

func checkInit(t *testing.T, stub *MockStub, args []string) {
  res := stub.MockInit("1", args)
  if res.Status != shim.OK {
    fmt.Println("Init failed", string(res.Message))
    t.FailNow()
  }
}

func checkState(t *testing.T, stub *MockStub, name string, value string) {
  bytes := stub.State[name]
  if bytes == nil {
    fmt.Println("State", name, "failed to get value")
    t.FailNow()
  }
  if string(bytes) != value {
    fmt.Println("State value", name, "was not", value, "as expected")
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

  attrVal, found, err := cid.GetAttributeValue(stub, "role")
  fmt.Printf("attrVal: %v, err: %v, found:%b \n", attrVal, err, found)
  // if err != nil {
  //   t.Errorf("Error getting Unique ID of the submitter of the transaction")
  // } else if !found {
  //   t.Errorf("Attribute 'role' should be found in the submitter cert")
  // } else if attrVal != "member" {
  //   t.Errorf("Assert have failed; value was %v, not member", attrVal)
  // }
}

// func Test_Init(t *testing.T) {

//   scc := new(SimpleChaincode)
//   stub := NewMockStub("loyalty", scc)
//   checkInit(t, stub, []string{"init", "A", "123", "B", "234"})

//   checkState(t, stub, "A", "123")
//   checkState(t, stub, "B", "234")
// }

func Test_Invoke(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := NewMockStub("loyalty", scc)
  checkInit(t, stub, []string{})

  checkInvoke(t, stub, "ping", "pong")
}
