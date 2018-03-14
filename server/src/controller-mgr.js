var defaultConfig = require("./config");
const controller_API = require("./controller");

const config = Object.assign({}, defaultConfig, {
  anotherUser: "admin",
  anotherUserSecret: "adminpw",
  user: "admin",
  MSP: defaultConfig.mspID
});

// console.log("Config:", config);

const controllerMap = new Map();
// user with channel will be unique, so that we do not have to re-initalized every time
// and these instances belong to the current organization
module.exports = {
  getInstance(user, channelName) {
    const key = user + "#" + channelName;
    if (!controllerMap.has(key)) {
      controllerMap.set(
        key,
        controller_API(Object.assign({}, config, { channelName: channelName }))
      );
    }
    return controllerMap.get(key);
  },
  getConfig() {
    return config;
  }
};
