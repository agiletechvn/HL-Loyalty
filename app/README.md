## Hyperledger Fabric Sample Application

Using dev mode
==============
Start network

```sh
./startFabric.sh
```

**Terminal 1 - Build & start the chaincode**
----------------------------------------

```sh
docker exec -it chaincode bash -c "cd sacc; go build -i && CORE_CHAINCODE_ID_NAME=mycc:1 ./sacc"
# after chaincode is changed, you can use combination of ctrl+c and start again
```

**Terminal 2 - Use the chaincode**
------------------------------

Even though you are in ``--peer-chaincodedev`` mode, you still have to install the
chaincode so the life-cycle system chaincode can go through its checks normally.
This requirement may be removed in future when in ``--peer-chaincodedev`` mode.

We'll leverage the CLI container to drive these calls.

```sh
# upgrade chaincode
docker exec dockercli peer chaincode install -n mycc -v 1.0 -p github.com/sacc
# instantiate chaincode
docker exec dockercli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n mycc -v 1.0 -c '{"Args":["a", "10"]}' 
# -P "OR ('Org1MSP.member','Org2MSP.member')"
```
