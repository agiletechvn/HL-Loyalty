const { authenticate } = require('./users/authenticate');
const { findUserCurrent } = require('./users/findUserCurrent');
const { findUsers } = require('./users/findUsers');
const { signUp } = require('./users/signUp');

const services = {
  users: {
    authenticate,
    findUserCurrent,
    findUsers,
    signUp
  }
};

module.exports = services;
