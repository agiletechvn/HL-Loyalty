/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/
/*
 * Chaincode query
 */
// var fs = require("fs-extra");
const x509 = require("x509");
const Fabric_Client = require("fabric-client");
const Fabric_Utils = require("fabric-client/lib/utils.js");
const path = require("path");
const util = require("util");
const fs = require("fs");
const os = require("os");
// const moment = require("moment");
const User = require("fabric-client/lib/User.js");
const CaService = require("fabric-ca-client/lib/FabricCAClientImpl.js");

module.exports = function(config) {
	const fabric_client = new Fabric_Client();
	const channel = fabric_client.newChannel(config.channelName);
	const grpcProtocol = config.tlsEnabled ? "grpcs://" : "grpc://";
	const peerConfig = config.tlsEnabled
		? {
				pem: config.peerPem,
				"ssl-target-name-override":
					config.peerDomain || config.peerHost.split(":")[0]
		  }
		: null;
	const peer = fabric_client.newPeer(
		grpcProtocol + config.peerHost,
		peerConfig
	);

	const ordererConfig = config.tlsEnabled
		? {
				pem: config.ordererPem,
				"ssl-target-name-override":
					config.ordererDomain || config.ordererHost.split(":")[0]
		  }
		: null;

	const orderer = fabric_client.newOrderer(
		grpcProtocol + config.ordererHost,
		ordererConfig
	);
	const store_path =
		process.env.KEY_STORE_PATH || path.join(os.homedir(), ".hfc-key-store");

	channel.addPeer(peer);
	channel.addOrderer(orderer);

	console.log("Peer: " + grpcProtocol + config.peerHost);
	console.log("Store path:" + store_path);

	// var logFile = process.env.NAMESPACE + '.' + config.channelName + '.csv'
	// console.log('logFile', logFile)
	let currentSubmitter = null;
	const instance = {
		viewca(user) {
			var obj = JSON.parse(fs.readFileSync(store_path + "/" + user, "utf8"));
			var cert = x509.parseCert(obj.enrollment.identity.certificate);
			return cert;
		},

		get_member_user(user) {
			// console.log(user);
			if (currentSubmitter) {
				// console.log("has found");
				return Promise.resolve(currentSubmitter);
			}

			// Fabric_Utils.setConfigSetting(
			//   "key-value-store",
			//   "fabric-client/lib/impl/CouchDBKeyValueStore.js"
			// );

			// const keyvalueStoreConfig = {
			//   name: "mychannel",
			//   url: "http://localhost:5984"
			// };

			const keyvalueStoreConfig = {
				path: store_path
			};

			return Fabric_Client.newDefaultKeyValueStore(keyvalueStoreConfig)
				.then(store => {
					fabric_client.setStateStore(store);
					const crypto_suite = Fabric_Client.newCryptoSuite();
					// use the same location for the state store (where the users' certificate are kept)
					// and the crypto store (where the users' keys are kept)
					const crypto_store = Fabric_Client.newCryptoKeyStore(
						keyvalueStoreConfig
					);
					crypto_suite.setCryptoKeyStore(crypto_store);
					fabric_client.setCryptoSuite(crypto_suite);
					return fabric_client.getUserContext(user, true);
				})
				.then(submitter => {
					if (submitter && submitter.isEnrolled()) {
						// console.log(submitter);
						currentSubmitter = submitter;
						return submitter;
					} else {
						throw new Error("failed");
					}
				})
				.catch(err => {
					console.error("[err]", err);
				});
		},

		getEventTxPromise(eventAdress, transaction_id_string) {
			// by default, orderer using batch timer about 2 seconds, after that it will do the batch cut and broadcast
			// to commiting peers
			return new Promise((resolve, reject) => {
				let event_hub = fabric_client.newEventHub();
				event_hub.setPeerAddr("grpc://" + eventAdress);
				console.log("eventhub: grpc://" + eventAdress);

				let handle = setTimeout(() => {
					event_hub.unregisterTxEvent(transaction_id_string);
					resolve(null); //we could use reject(new Error('Trnasaction did not complete within 30 seconds'));
				}, 150000);
				// register everything
				event_hub.connect();
				event_hub.registerTxEvent(
					transaction_id_string,
					(event_tx_id, status) => {
						clearTimeout(handle);
						//channel_event_hub.unregisterTxEvent(event_tx_id); let the default do this
						console.log("Successfully received the transaction event");
						event_hub.unregisterTxEvent(transaction_id_string);
						resolve(transaction_id_string);
					},
					error => {
						reject(
							new Error("There was a problem with the eventhub ::" + error)
						);
					}
				);
			});
		},

		query(user, request) {
			return this.get_member_user(user)
				.then(user_from_store => {
					return channel.queryByChaincode(request);
				})
				.then(query_responses => {
					// query_responses could have more than one  results if there multiple peers were used as targets
					if (query_responses && query_responses.length == 1) {
						if (query_responses[0] instanceof Error) {
							throw query_responses[0];
						} else {
							return query_responses[0];
						}
					} else {
						console.log("No payloads were returned from query");
						return null;
					}
				});
		},

		invoke(user, invokeRequest) {
			var tx_id;

			return this.get_member_user(user)
				.then(user_from_store => {
					tx_id = fabric_client.newTransactionID();

					console.log("invokeRequest:", invokeRequest);

					return channel.sendTransactionProposal({
						chaincodeId: invokeRequest.chaincodeId,
						fcn: invokeRequest.fcn,
						args: invokeRequest.args,
						txId: tx_id
					});
				})
				.then(results => {
					var proposalResponses = results[0];
					var proposal = results[1];

					if (
						proposalResponses &&
						proposalResponses[0].response &&
						proposalResponses[0].response.status === 200
					) {
						var transaction_id_string = tx_id.getTransactionID();
						const txPromise = this.getEventTxPromise(
							config.eventHost,
							transaction_id_string
						);

						const sendPromise = channel.sendTransaction({
							proposalResponses: proposalResponses,
							proposal: proposal
						});

						return Promise.all([sendPromise, txPromise]);
					} else {
						throw proposalResponses[0];
					}
				});
		}
	};

	return instance;
};
