#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
# This code is based on code written by the Hyperledger Fabric community. 
# Original code can be found here: https://github.com/hyperledger/fabric-samples/blob/release/fabcar/startFabric.sh
#
# Exit on first error

set -e

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1

starttime=$(date +%s)

if [ ! -d ~/.hfc-key-store/ ]; then
	mkdir ~/.hfc-key-store/
fi

# launch network; create channel and join peer to channel
cd ../basic-network
./start.sh

# Now launch the dockercli container in order to install, instantiate chaincode
# and prime the ledger with our 10 tuna catches
# docker-compose -f ./docker-compose.yml up -d dockercli

echo "Cleaning chaincode images and container..."
echo
# Delete docker containers
dockerContainers=$(docker ps -a --format '{{.ID}} {{.Names}}' | awk '$2~/^dev-peer/{print $1}')
if [ "$dockerContainers" != "" ]; then     
  docker rm -f $dockerContainers > /dev/null
fi

chaincodeImages=$(docker images --format '{{.ID}} {{.Repository}}' | awk '$2~/^dev-peer/{print $1}')  
if [ "$chaincodeImages" != "" ]; then     
  docker rmi $chaincodeImages > /dev/null
fi  

# docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" dockercli peer chaincode install -n tuna-app -v 1.0 -p github.com/tuna-app
# docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" dockercli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n tuna-app -v 1.0 -c '{"Args":[""]}' -P "OR ('Org1MSP.member','Org2MSP.member')"
# sleep 3
# docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" dockercli peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n tuna-app -c '{"function":"initLedger","Args":[""]}'

# install read/write chaincode
# docker exec -it chaincode bash -c "cd sacc; go build -i && CORE_CHAINCODE_ID_NAME=mycc:1.0 ./sacc"
# docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" dockercli peer chaincode install -n mycc -v 1.0 -p github.com/sacc
# docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" dockercli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n mycc -v 1.0 -c '{"Args":["a", "10"]}' # -P "OR ('Org1MSP.member','Org2MSP.member')"
# sleep 3
# docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" dockercli peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n sacc -c '{"function":"set","Args":["a","20"]}'
# docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" dockercli peer chaincode query -o orderer.example.com:7050 -C mychannel -n sacc -c '{"function":"query","Args":["a"]}'

printf "\nTotal execution time : $(($(date +%s) - starttime)) secs ...\n\n"
printf "\nStart with the registerAdmin.js, then registerUser.js, then server.js\n\n"

cd ../tuna-app
rm -rf ~/.hfc-key-store/*
node registerAdmin.js
# node registerUser.js
