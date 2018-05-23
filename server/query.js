'use strict';
/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/
/*
 * Chaincode query
 */
var fs = require('fs-extra');
var x509 = require('x509');
var Fabric_Client = require('fabric-client');
var path = require('path');
var util = require('util');
var os = require('os');
var program = require('commander');

//
var fabric_client = new Fabric_Client();

program
  .version('0.1.0')
  .option('-u, --user []', 'User id', 'user1')
  .option('--channel, --channel []', 'A channel', 'mychannel')
  .option('--chaincode, --chaincode []', 'A chaincode', 'origincert')
  .option('-m, --method []', 'A method', 'getCreator')
  .option(
    '-a, --arguments [value]',
    'A repeatable value',
    (val, memo) => memo.push(val) && memo,
    []
  )
  .parse(process.argv);

// setup the fabric network
var channel = fabric_client.newChannel(program.channel);
var peer = fabric_client.newPeer('grpc://localhost:7051');
channel.addPeer(peer);
//
var member_user = null;
var store_path = path.join(__dirname, 'hfc-key-store');
console.log('Store path:' + store_path);
var tx_id = null;

Fabric_Client.newDefaultKeyValueStore({
  path: store_path
})
  .then(state_store => {
    // assign the store to the fabric client
    fabric_client.setStateStore(state_store);
    var crypto_suite = Fabric_Client.newCryptoSuite();
    // use the same location for the state store (where the users' certificate are kept)
    // and the crypto store (where the users' keys are kept)
    var crypto_store = Fabric_Client.newCryptoKeyStore({ path: store_path });
    crypto_suite.setCryptoKeyStore(crypto_store);
    fabric_client.setCryptoSuite(crypto_suite);

    // get the enrolled user from persistence, this user will sign all requests
    return fabric_client.getUserContext(program.user, true);
  })
  .then(user_from_store => {
    if (user_from_store && user_from_store.isEnrolled()) {
      console.log('Successfully loaded ' + program.user + ' from persistence');
      member_user = user_from_store;
    } else {
      throw new Error(
        'Failed to get ' +
          program.user +
          '.... run node register.js -u ' +
          program.user
      );
    }

    // queryCar chaincode function - requires 1 argument, ex: args: ['CAR4'],
    // queryAllCars chaincode function - requires no arguments , ex: args: [''],
    const request = {
      //targets : --- letting this default to the peers assigned to the channel
      chaincodeId: program.chaincode,
      fcn: program.method,
      args: program.arguments
    };
    return channel.queryByChaincode(request);
  })
  .then(query_responses => {
    console.log(
      'Query has completed on channel [' +
        program.channel +
        '], checking results'
    );
    // query_responses could have more than one  results if there multiple peers were used as targets
    if (query_responses && query_responses.length == 1) {
      if (query_responses[0] instanceof Error) {
        console.error('error from query = ', query_responses[0]);
      } else {
        const response = parseResponse(query_responses);
        console.log('Response is \n', response);
      }
    } else {
      console.log('No payloads were returned from query');
    }
  })
  .catch(err => {
    console.error('Failed to query successfully :: ' + err);
  });

function parseResponse(response) {
  return response[0].toString();
  // const parsedJson = JSON.parse(response[0].toString());
  // parsedJson.id_bytes = Buffer.from(parsedJson.id_bytes, "base64").toString(
  //   "utf8"
  // );
  // const cert = x509.parseCert(parsedJson.id_bytes);
  // parsedJson.attrs = cert.extensions["1.2.3.4.5.6.7.8.1"];
  // return parsedJson;
}
