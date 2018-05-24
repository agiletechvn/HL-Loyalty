cd $GOPATH/src/loyalty
nodemon --exec "./startChaincode.sh -v $2 -a $1" loyalty.go