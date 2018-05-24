## Hyperledger Fabric Sample Application

## **Install the requirements**

https://github.com/Masterminds/glide

```sh
glide install
```

## **Run the test**

```sh
cd ./chaincode/loyalty
ln -s $PWD $GOPATH/src/
cd $GOPATH/src/loyalty
CGO_LDFLAGS_ALLOW='-Wl,--no-as-needed' go test
```

## **Start chaincode**

```sh
# cd ./chaincode/loyalty
# nodemon --exec "./startChaincode.sh" loyalty.go
yarn start
# run at background
yarn start > /dev/null 2>&1 &
# run on kubernetes
yarn start-k8s 1.0
```

## **Install on kubernetes**

```sh
cp chaincode/loyalty/ $share_folder/channel-artifacts/chaincode/loyalty
# then later use k8s script to setup
```
