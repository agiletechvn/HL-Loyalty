chaincodeID=$(ps -ax | awk '$0~/\.\/loyalty/{print $1}')
if [ ! -z "$chaincodeID" ];then  
  kill $chaincodeID
fi

while getopts "a:v:" opt; do
  case "$opt" in
    a)  ADDRESS=$OPTARG
    ;;
    v)  VERSION=$OPTARG
    ;;
    *) 
      echo "Unknown $opt"
      exit 1
    ;;
  esac
done

: ${VERSION:=1.0}
: ${ADDRESS:=localhost:7052}

chmod u+x ./loyalty
# sleep 1
echo "go build -i && CORE_PEER_ADDRESS=$ADDRESS CORE_PEER_LOCALMSPID=Org1MSP CORE_VM_ENDPOINT=unix:///var/run/docker.sock CORE_CHAINCODE_ID_NAME=mycc:$VERSION ./loyalty"
go build -i && CORE_PEER_ADDRESS=$ADDRESS CORE_PEER_LOCALMSPID=Org1MSP CORE_VM_ENDPOINT=unix:///var/run/docker.sock CORE_CHAINCODE_ID_NAME=mycc:$VERSION ./loyalty
