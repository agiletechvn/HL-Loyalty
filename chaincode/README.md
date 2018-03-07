## Hyperledger Fabric Sample Application


**Install the requirements**
----------------------------------------
https://github.com/Masterminds/glide
```sh
glide install
```

**Run the test**
----------------------------------------

```sh
cd ./chaincode/loyalty
ln -s $PWD $GOPATH/src/
cd $PWD $GOPATH/src/loyalty
go test
```

