## Hyperledger Fabric Sample Application

# Using dev mode

Start network

```sh
# start network
./startFabric.sh
# start api
yarn start
```

Register new user

```sh
yarn enroll admin
yarn register user1
yarn enroll user1
```

## **Terminal 1 - Build & start the chaincode**

```sh
cd ./chaincode
# start chaincode
yarn start 1.0
# after chaincode is changed, you can use combination of ctrl+c and start again
```

## **Terminal 2 - Use the chaincode**

Even though you are in `--peer-chaincodedev` mode, you still have to install the
chaincode so the life-cycle system chaincode can go through its checks normally.
This requirement may be removed in future when in `--peer-chaincodedev` mode.

We'll leverage the CLI container to drive these calls.

```sh
# upgrade chaincode
docker exec cli peer chaincode install -n mycc -v 1 -p github.com/chaincode/loyalty
# instantiate chaincode
docker exec cli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n mycc -v 1 -c '{"Args":[]}'
# -P "OR ('Org1MSP.member','Org2MSP.member')"

# query chaincode
docker exec cli peer chaincode query -n mycc -C mychannel -c '{"Args":["ping"]}'
```
