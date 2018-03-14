//SPDX-License-Identifier: Apache-2.0

// nodejs server setup

// call the packages we need
const express = require("express"); // call express
const cors = require("cors");
const bodyParser = require("body-parser");
const http = require("http");
const fs = require("fs");
const Fabric_Client = require("fabric-client");
const path = require("path");
const util = require("util");
const os = require("os");

const publicPath = path.join(__dirname, "../client/build");

const app = express(); // define our app using express
// Load all of our middleware
// configure app to use bodyParser()
// this will let us get the data from a POST
// app.use(express.static(__dirname + '/client'));
app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());
app.use(cors());
app.use("/api", require("./src/routes.js"));

// set up a static file server that points to the "client" directory
app.use(express.static(publicPath));
app.get("*", (req, res) => {
  res.sendFile(path.resolve(publicPath + "/index.html"));
});

// Save our port
const port = process.env.PORT || 8000;

// Start the server and listen on port
app.listen(port, function() {
  console.log("Live on port: " + port);
});
