#!/usr/bin/env bash
set -e
echo "Install envroiment"
yum install epel-release -y

echo "install go"
GO_ROOT=/data/install/
echo "install golang 1.10.2..."
sudo rpm --import https://mirror.go-repo.io/centos/RPM-GPG-KEY-GO-REPO
curl -s https://mirror.go-repo.io/centos/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo
sudo yum install golang -y
echo "install golang 1.10.2 success."

sudo mkdir -p $GO_ROOT/go/{bin,pkg,src}

export GO_ROOT=/data/install
export GOPATH=$GO_ROOT/go
export PATH="$PATH:${GOPATH//://bin:}/bin"

echo 'export GO_ROOT=/data/install' >> ~/.bashrc
echo 'export GOPATH=$GO_ROOT/go' >> ~/.bashrc
echo 'export PATH="$PATH:${GOPATH//://bin:}/bin"' >> ~/.bashrc

echo "install docker-ce"
sudo yum install -y yum-utils device-mapper-persistent-data lvm2
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum makecache fast
sudo yum install docker-ce -y
echo "install docker-compose"
sudo curl -L https://github.com/docker/compose/releases/download/1.21.2/docker-compose-$(uname -s)-$(uname -m) -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

sudo yum install epel-release -y
sudo yum install wget python-pip python-devel libtool-ltdl libtool-ltdl-devel -y

echo "Install git"
sudo yum install git -y

cd $GOPATH
mkdir -p src/github.com/hyperledger
cd src/github.com/hyperledger
git clone https://github.com/hyperledger/fabric.git

echo "install yaml.v2"
go get gopkg.in/yaml.v2
cd $GOPATH/src/github.com/hyperledger/fabric/
echo "build crypto tools"
make configtxgen
make cryptogen
make configtxlator
sudo cp build/bin/* /usr/local/bin


echo "Install envroiment successful"

