package main

import (
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/hyperledger/fabric/protos/peer"
)

type MockStub struct {
  *shim.MockStub
  creator []byte
  cc      shim.Chaincode
  args    []string
}

func NewMockStub(name string, cc shim.Chaincode) *MockStub {
  return &MockStub{
    shim.NewMockStub(name, cc),
    nil,
    cc,
    nil,
  }
}

func (stub *MockStub) GetStringArgs() []string {
  return stub.args
}

// use name from return so that we just assign those, no need to return
func (stub *MockStub) GetFunctionAndParameters() (string, []string) {
  return stub.args[0], stub.args[1:]
}

func (stub *MockStub) GetCreator() ([]byte, error) {
  return stub.creator, nil
}

func (stub *MockStub) GetChaincode() *SimpleChaincode {
  return stub.cc.(*SimpleChaincode)
}

func (stub *MockStub) MockInit(uuid string, args []string) peer.Response {
  stub.args = args
  stub.MockTransactionStart(uuid)
  res := stub.GetChaincode().Init(stub)
  stub.MockTransactionEnd(uuid)
  return res
}

func (stub *MockStub) MockInvoke(uuid string, args []string) peer.Response {
  stub.args = args
  stub.MockTransactionStart(uuid)
  res := stub.GetChaincode().Invoke(stub)
  stub.MockTransactionEnd(uuid)
  return res
}
