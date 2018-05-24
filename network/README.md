## Basic Network Config

Note that this basic configuration uses pre-generated certificates and
key material, and also has predefined transactions to initialize a
channel named "mychannel".

To regenerate this material, simply run `generate.sh`.

To start the network, run `start.sh` or `yarn start`.
To stop it, run `stop.sh` or `yarn stop`
To completely remove all incriminating evidence of the network
on your system, run `teardown.sh` or `yarn teardown`.
To reset the network, run `yarn reset`.

## Install chaincode

```bash
yarn installChaincode 1.0
# instantiate
yarn instantiateChaincode 1.0
```


