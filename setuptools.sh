BASE_DIR=$PWD/network

verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
                echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
		echo
   		exit 1
	fi
}

cd $GOPATH/src
mkdir -p github.com/hyperledger
cd github.com/hyperledger
git clone https://github.com/hyperledger/fabric.git
go get gopkg.in/yaml.v2
cd $GOPATH/src/github.com/hyperledger/fabric/
CGO_LDFLAGS_ALLOW="-I.*" make configtxgen
res=$?
CGO_LDFLAGS_ALLOW="-I.*" make cryptogen  
((res+=$?))
CGO_LDFLAGS_ALLOW="-I.*" make configtxlator  
((res+=$?))
verifyResult $res "Build crypto tools failed"
echo "===================== Crypto tools built successfully ===================== "
echo 
echo "Copying to bin folder of network..."
echo
mkdir -p ${BASE_DIR}/bin/
cp ./build/bin/configtxgen ${BASE_DIR}/bin/
cp ./build/bin/cryptogen ${BASE_DIR}/bin/
cp ./build/bin/configtxlator ${BASE_DIR}/bin/