cd $GOPATH/src/loyalty
nodemon --exec "./startChaincode.sh -v ${2:-1.0} -a $1" loyalty.go