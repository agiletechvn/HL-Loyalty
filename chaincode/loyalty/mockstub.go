package main

import (
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/hyperledger/fabric/protos/peer"
)

type MockStub struct {
  *shim.MockStub
  creator []byte
}

func NewMockStub(name string, cc shim.Chaincode) *MockStub {
  return &MockStub{
    shim.NewMockStub(name, cc),
    nil,
  }
}

func StringArgsToBytesArgs(args []string) [][]byte {
  a := make([][]byte, len(args))
  for k, v := range args {
    a[k] = []byte(v)
  }
  return a
}

func (m *MockStub) GetCreator() ([]byte, error) {
  return m.creator, nil
}

func (m *MockStub) MockInit(uuid string, args []string) peer.Response {
  return m.MockStub.MockInit(uuid, StringArgsToBytesArgs(args))
}

func (m *MockStub) MockInvoke(uuid string, args []string) peer.Response {
  return m.MockStub.MockInvoke(uuid, StringArgsToBytesArgs(args))
}
