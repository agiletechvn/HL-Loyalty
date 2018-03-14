//SPDX-License-Identifier: Apache-2.0
// nodejs server setup
const router = require("express").Router();
// call the packages we need
// const path = require("path");
// const fs = require("fs");
// const os = require("os");

const controllerManager = require("./controller-mgr");
const config = controllerManager.getConfig();

router.get("/viewca", function(req, res) {
  const user = req.query.user || config.user;
  const controller = controllerManager.getInstance(
    user,
    req.query.channel || "mychannel"
  );

  // each method require different certificate of user
  const ret = controller.viewca(user);
  res.send(ret);
});

router.get("/query", function(req, res) {
  const user = req.query.user || config.user;
  const controller = controllerManager.getInstance(
    user,
    req.query.channel || "mychannel"
  );
  const request = {
    //targets : --- letting this default to the peers assigned to the channel
    chaincodeId: req.query.chaincode,
    fcn: req.query.method,
    args: req.query.arguments || []
  };

  // each method require different certificate of user
  controller
    .query(user, request)
    .then(ret => {
      res.send(ret.toString());
    })
    .catch(err => {
      res.status(500).send(err);
    });
});

router.get("/invoke", function(req, res) {
  const user = req.query.user || config.user;
  const controller = controllerManager.getInstance(
    user,
    req.query.channel || "mychannel"
  );
  const request = {
    chaincodeId: req.query.chaincode,
    fcn: req.query.method,
    args: req.query.arguments || []
  };
  // each method require different certificate of user
  controller
    .invoke(req.query.user || config.user, request)
    .then(ret => {
      res.json(ret);
    })
    .catch(err => {
      console.log(err);
      res.status(500).send(err);
    });
});

module.exports = router;
