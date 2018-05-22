chaincodeID=$(ps -ax | awk '$0~/\.\/loyalty/{print $1}')
if [[ ! -z $chaincodeID ]];then  
  kill $chaincodeID
fi
# sleep 1
go build -i && CORE_PEER_ADDRESS=localhost:7052 CORE_PEER_LOCALMSPID=Org1MSP CORE_VM_ENDPOINT=unix:///var/run/docker.sock CORE_CHAINCODE_ID_NAME=mycc:$1 ./loyalty