## Hyperledger Fabric Loyalty System

## **Install the requirements**

On Centos you can run this [script](install_fabric_centos.sh) for pre-setup

Requirement Modules (You can download packages on websites if using desktop):

*   Nodejs, npm, yarn, nodemon

```sh
curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash - && sudo apt-get install -y nodejs
curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
sudo apt-get update && sudo apt-get install yarn
yarn global add nodemon
```

*   Go lang, glide

If "go version" > 1.8 -> no need for update
If < 1.8 :

```sh
apt remove golang-go -y
rm -rf /usr/local/go
wget https://dl.google.com/go/go1.10.1.linux-amd64.tar.gz
sudo tar -xvf go1.10.1.linux-amd64.tar.gz
sudo mv go /usr/local
nano ~/. profile
```

Just copy below lines:

```sh
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

Then

```sh
source ~/. profile
nano ~/.bashrc
```

Copy: export GOPATH=/opt/gopath

```sh
source ~/. bashrc
```

Checking the version again: go version

```sh
curl https://glide.sh/get | sh
```

If error is like "[...]/src not found":

```sh
mkdir $GOPATH/src
```

Then run again

*   docker, docker-compose (You can download docker GUI instead of command line)

```sh
sudo apt install docker.io
sudo systemctl start docker
sudo systemctl enable docker
sudo curl -L https://github.com/docker/compose/releases/download/1.21.2/docker-compose-$(uname -s)-$(uname -m) -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

*   python, pip

```sh
sudo apt-get install python
sudo apt-get install python-setuptools -y
sudo easy_install pip
```

## Other tools might be needed for installing

If some errors are like "Failed at the pkcs11js" or "script 'node-gyp rebuild'", then do:

```sh
npm install -g node-gyp
npm config set python /usr/bin/python
yum install gcc-c++
sudo apt install build-essential g++
sudo apt install libtool libltdl-dev
```

## Step0: Install Fabric tools:

```sh
./setuptools.sh
```

## Step1: Setup the network

```sh
cd network/
yarn
yarn generate
yarn start
yarn installChaincode 1.0
```

(1.0 is the version of chaincode)

Optional:

```sh
yarn stop (to stopp current app)
yarn teardown (to kill and delete all)
yarn reset (to reset app)
```

## Step2: Setup the chaincode

```sh
cd ../chaincode/
cd loyalty && glide install
ln -s $PWD $GOPATH/src/
yarn start
```

```sh
cd ../../network
yarn instantiateChaincode 1.0
```

## Step3: Setup and run the api and client

```sh
cd ../server && yarn
rm -rf hfc-key-store/*
yarn enroll admin
```

(we just need to enroll admin, because we created "admin" by the code)
Next, we can create a user for starting querying and making some transactions

```sh
yarn register user1
yarn enroll user1
yarn start
```

(server runs at 8000 port)

```sh
cd ../client && yarn && yarn start (client-dev runs at 3000 port)
```

**[Network](network/README.md)**  
**[Chaincode](chaincode/README.md)**  
**[Server](server/README.md)**
