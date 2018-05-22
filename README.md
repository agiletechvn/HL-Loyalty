## Hyperledger Fabric Sample Application

## **Install the requirements**

**Install nodemon**

```sh
yarn global add nodemon
```

**Start server & network**

```sh
cd server
./startFabric.sh
yarn start
```

**Start chaincode**

```sh
cd chaincode
yarn start
```

**Instantiate, Upgrade chaincode**

```sh
docker exec -it dockercli bash
peer chaincode install -n mycc -v 1.0 -p github.com/chaincode/loyalty
peer chaincode instantiate -C mychannel -o orderer.example.com:7050 -n mycc -v 1.0 -c '{"Args":[]}'
peer chaincode upgrade -C mychannel -o orderer.example.com:7050 -n mycc -v 2.0 -c '{"Args":[]}'
```
