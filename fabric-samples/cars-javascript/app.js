/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
const express = require('express');
const app = express();
//var contract = {};

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const bodyParser = require('body-parser');
const path = require('path');
const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('../test-application/javascript/CAUtil.js');
const { buildCCPOrg1, buildWallet } = require('../test-application/javascript/AppUtil.js');

const swaggerUi = require('swagger-ui-express');
const swaggerJSDoc = require('swagger-jsdoc');
const swaggerAutogen = require('swagger-autogen');
const outputFile = './swagger_output.json'
const endpointsFiles = ['./endpoints.js']
const fs = require('fs');
swaggerAutogen(outputFile, endpointsFiles).then(() => {
    require('./app.js')
})
const options = {
  swaggerDefinition: {
    info: {
      title: 'My API',
      version: '1.0.0',
      description: 'My API documentation',
    },
    basePath: '/',
  },
  apis: ['path/to/my/routes/*.js'],
};

const swaggerSpec = swaggerJSDoc(options);
const jsonParser = bodyParser.json();

const channelName = 'mychannel';
const chaincodeName = 'basic';
const mspOrg1 = 'Org1MSP';
const walletPath = path.join(__dirname, 'wallet');
const org1UserId = 'appUser';

function prettyJSONString(inputString) {
	return JSON.stringify(JSON.parse(inputString), null, 2);
}

function getConfig() {
	const configFile = 'config.json';
  
	// Read configuration from file
	try {
	  const configData = fs.readFileSync(configFile);
	  const config = JSON.parse(configData);
	  return config;
	} catch (err) {
	  console.error(`Failed to read configuration file: ${err.message}`);
	  return null;
	}
  }
  
  module.exports = getConfig;
  

app.get('/', (req, res) => {
	//getConfig();
	res.send('Hello, World!');
	});

// Start the server
app.listen(3000, () => {
	console.log('Server is listening on port 3000');
	main();
});
app.use("/api-docs", swaggerUi.serve, swaggerUi.setup(swaggerSpec));

//ledger init
app.post('/initLedger', (req, res) => {
	initLedger(contract);
	res.send(`Initialized ledger.`);
});

//read person asset
app.post('/readPersonAsset', (req, res) => {
	// Extract parameters from request body
	const { id } = req.body;
	readPersonAsset(contract, id);
	// Send response back to the client
	res.send(`Transaction completed.`);
});

//read car asset
app.post('/readCarAsset', (req, res) => {
	// Extract parameters from request body
	const { id } = req.body;
	readCarAsset(contract, id);
	// Send response back to the client
	res.send(`Transaction completed.`);
});

//add car malfunction
app.post('/addCarMalfunction', jsonParser, (req, res) => {
	// Extract parameters from request body
	console.log(req.body);
	const { id, description, repairPrice } = req.body;
	// Call addCarMalfunction function with parameters
	addCarMalfunction(contract, id, description, repairPrice);
	// Send response back to the client
	res.send(`Added car malfunction for car ${id}`);
});

//get cars by color
app.get('/getCarsByColor/:color', (req, res) => {
	// Extract parameters from request body
	const color = req.params.color;
	getCarsByColor(contract, color);
	// Send response back to the client
	res.send(`Got cars of color: ${color}.`);
});

app.get('/getCarsByColorAndOwner', (req, res) => {
	const { color, owner } = req.query;
	getCarsByColorAndOwner(contract, color, owner);
	res.send(`Getting cars of color ${color}, owner id ${owner}`)
});

app.post('/transferCarAsset', (req, res) => {
	const { id, newOwner, acceptMalfunction } = req.body;
	transferCarAsset(contract, id, newOwner, acceptMalfunction);
	res.send(`Transfering car ${id}, to owner id ${newOwner}`);
});

app.post('/changeCarColor', (req, res) => {
	const { id, newColor } = req.body;
	changeCarColor(contract, id, newColor);
	res.send(`Changing car color ${id}, to  ${newColor}`);
});

app.post('/repairCar', (req, res) => {
	const { id } = req.body;
	repairCar(contract, id);
	res.send(`Repairing car ${id}`);
});


async function main() {
	try {
		// build an in memory object with the network configuration (also known as a connection profile)
		const ccp = buildCCPOrg1();

		// build an instance of the fabric ca services client based on
		// the information in the network configuration
		const caClient = buildCAClient(FabricCAServices, ccp, 'ca.org1.example.com');
		
		// setup the wallet to hold the credentials of the application user
		const wallet = await buildWallet(Wallets, walletPath);

		// in a real application this would be done on an administrative flow, and only once
		//await enrollAdmin(caClient, wallet, mspOrg1);

		// in a real application this would be done only when a new user was required to be added
		// and would be part of an administrative flow
		//await registerAndEnrollUser(caClient, wallet, mspOrg1, org1UserId, 'org1.department1');

		// Create a new gateway instance for interacting with the fabric network.
		// In a real application this would be done as the backend server session is setup for
		// a user that has been verified.
		const gateway = new Gateway();

		try {
			// setup the gateway instance
			// The user will now be able to create connections to the fabric network and be able to
			// submit transactions and query. All transactions submitted by this gateway will be
			// signed by this user using the credentials stored in the wallet.
			await gateway.connect(ccp, {
				wallet,
				identity: org1UserId,
				discovery: { enabled: true, asLocalhost: true } // using asLocalhost as this gateway is using a fabric network deployed locally
			});

			// Build a network instance based on the channel where the smart contract is deployed
			const network = await gateway.getNetwork(channelName);

			// Get the contract from the network.
			global.contract = network.getContract(chaincodeName);
		} finally {
			// Disconnect from the gateway when the application is closing
			// This will close all connections to the network
			gateway.disconnect();
		}
	} catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
	}
}	

// newGrpcConnection creates a gRPC connection to the Gateway server.
function newGrpcConnection(tlsCertPath, gatewayPeer, peerEndpoint) {
  const packageDefinition = protoLoader.loadSync(
    path.resolve(__dirname, './protos/gateway.proto'),
    {keepCase: true, longs: String, enums: String, defaults: true, oneofs: true
  });
  const gatewayProto = grpc.loadPackageDefinition(packageDefinition).gateway;

  const certificate = fs.readFileSync(tlsCertPath);
  const certPool = grpc.credentials.createSsl(certificate);
  const credentials = grpc.credentials.combineChannelCredentials(certPool);

  return new gatewayProto.Gateway(gatewayPeer, credentials).connect(peerEndpoint);
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
function newIdentity(certPath, mspID) {
  const certificate = fs.readFileSync(certPath);
  return new identity(mspID, certificate);
}

function loadCertificate(filename) {
  const certificatePEM = fs.readFileSync(filename);
  return identity.CertificateFromPEM(certificatePEM);
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
function newSign(keyPath) {
  const privateKeyPath = fs.readdirSync(keyPath)[0];
  const privateKeyPEM = fs.readFileSync(path.join(keyPath, privateKeyPath));
  const privateKey = identity.PrivateKeyFromPEM(privateKeyPEM);
  return privateKey.sign.bind(privateKey);
}

/*
 This type of transaction would typically only be run once by an application the first time it was started after its
 initial deployment. A new version of the chaincode deployed later would likely not need to run an "init" function.
*/
async function initLedger(contract) {
  console.log('Submit Transaction: InitLedger, function creates the initial set of assets on the ledger');

  try {
    await contract.submitTransaction('InitLedger');
    console.log('*** Transaction committed successfully');
  } catch (err) {
    console.error(`failed to submit transaction: ${err}`);
  }
}

async function readPersonAsset(contract, id) {
  console.log('Evaluate Transaction: ReadPersonAsset, function returns person asset attributes');

  try {
    const evaluateResult = await contract.evaluateTransaction('ReadPersonAsset', id);
    const result = JSON.stringify(evaluateResult);

    console.log(`*** Result: ${result}`);
  } catch (err) {
    console.error(`failed to evaluate transaction: ${err}`);
  }
}

async function readCarAsset(contract, id) {
  console.log('Evaluate Transaction: ReadCarAsset, function returns car asset attributes');

  try {
    const evaluateResult = await contract.evaluateTransaction('ReadCarAsset', id);
    const result = JSON.stringify(evaluateResult);

    console.log(`*** Result: ${result}`);
  } catch (err) {
    console.error(`failed to evaluate transaction: ${err}`);
  }
}

async function getCarsByColor(contract, color) {
	console.log("Evaluate Transaction: GetCarsByColor, function returns all the cars with the given color");
  
	try {
	  const evaluateResult = await contract.evaluateTransaction("GetCarsByColor", color);
	  const result = formatJSON(evaluateResult);
  
	  console.log(`*** Result:${result}\n`);
	} catch (err) {
	  console.log(`failed to evaluate transaction: ${err}`);
	}
}

async function getCarsByColorAndOwner(contract, color, ownerID) {
	console.log("Evaluate Transaction: GetCarsByColorAndOwner, function returns all the cars with the given color and owner");
  
	try {
	  const evaluateResult = await contract.evaluateTransaction("GetCarsByColorAndOwner", color, ownerID);
	  const result = formatJSON(evaluateResult);
	  console.log(`*** Result:${result}`);
	} catch (err) {
	  console.log(`failed to evaluate transaction: ${err}`);
	}
}
  
async function transferCarAsset(contract, id, newOwner, acceptMalfunction) {
	console.log("Submit Transaction: TransferCarAsset, change car owner");

	try {
		await contract.submitTransaction("TransferCarAsset", id, newOwner, acceptMalfunction.toString());
		console.log("*** Transaction committed successfully");
	} catch (err) {
		console.log(`failed to submit transaction: ${err}`);
	}
}

function addCarMalfunction(contract, id, description, repairPrice) {
	console.log("Submit Transaction: AddCarMalfunction, record a new car malfunction");
  
	contract.submitTransaction("AddCarMalfunction", id, description, repairPrice.toString())
	  .then(() => {
		console.log("*** Transaction committed successfully");
	  })
	  .catch((err) => {
		console.log("failed to submit transaction: " + err);
	  });
}
  
function changeCarColor(contract, id, newColor) {
	console.log("Submit Transaction: ChangeCarColor, change the color of a car");

	contract.submitTransaction("ChangeCarColor", id, newColor)
		.then(() => {
			console.log("*** Transaction committed successfully");
		})
		.catch((err) => {
			console.log("failed to submit transaction: " + err);
		});
}

function repairCar(contract, id) {
	console.log("Submit Transaction: RepairCar, fix all of the car's malfunctions");
  
	contract.submitTransaction("RepairCar", id)
	  .then(() => {
		console.log("*** Transaction committed successfully");
	  })
	  .catch((err) => {
		console.log("failed to submit transaction: " + err);
	  });
  }
  
  // Format JSON data
function formatJSON(data) {
	try {
		return JSON.stringify(JSON.parse(data), null, " ");
	} catch (err) {
		throw new Error("failed to parse JSON: " + err);
	}
}
  
  
  
  