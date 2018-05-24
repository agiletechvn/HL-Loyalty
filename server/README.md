## Hyperledger Fabric Sample Application

# Using dev mode

Start network

```sh
# start network
./startFabric.sh
# start api
yarn start
```

Register new user

```sh
yarn enroll admin
yarn register user1
yarn enroll user1
```

## **Run on kubernetes**

Start network

```sh
# start network
./startFabric.sh
# start api
yarn start-k8s
```

Register new user

```sh
yarn enroll-k8s admin
yarn register-k8s user1
yarn enroll-k8s user1
```
