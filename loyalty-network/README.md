Using dev mode
==============
Start network

```sh
docker-compose up -d
```

Terminal 1 - Build & start the chaincode
----------------------------------------

```sh
docker exec -it chaincode bash -c "cd sacc; go build -i && CORE_CHAINCODE_ID_NAME=mycc:1 ./sacc"
# after chaincode is changed, you can use combination of ctrl+c and start again
```

Terminal 2 - Use the chaincode
------------------------------

Even though you are in ``--peer-chaincodedev`` mode, you still have to install the
chaincode so the life-cycle system chaincode can go through its checks normally.
This requirement may be removed in future when in ``--peer-chaincodedev`` mode.

We'll leverage the CLI container to drive these calls.

```sh
docker exec -it cli bash
```

# instantiate
```sh
CC_VERSION=1; peer chaincode install -p chaincodedev/chaincode/sacc -n mycc -v $CC_VERSION; peer chaincode instantiate -n mycc -c '{"Args":["a","10"]}' -C myc -v $CC_VERSION
```

# upgrade
```sh
CC_VERSION=4; peer chaincode install -p chaincodedev/chaincode/sacc -n mycc -v $CC_VERSION; peer chaincode upgrade -n mycc -c '{"Args":["a","10"]}' -C myc -v $CC_VERSION
```
