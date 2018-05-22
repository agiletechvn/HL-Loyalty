version=${1:-"1.0"}
cd $GOPATH/src/loyalty
nodemon --exec "./startChaincode.sh $version" loyalty.go