const services = require('../services')
const { SuccessResponse } = require('../helpers/responseHelpers')

exports.getUsers = async (req, res) => {
  const users = await services.users.findUsers()

  return res.json(
    new SuccessResponse.Builder()
      .withContent(users)
      .build()
  )
}

exports.postAuthenticate = async (req, res) => {
  // Get input data
  let email = req.body.email
  let password = req.body.password

  let token = await services.users.authenticate(email, password)

  return res.json(
    new SuccessResponse.Builder()
      .withContent(token)
  )
}

exports.postSignUp = async (req, res) => {
  // Get input data
  let name = req.body.name
  let email = req.body.email
  let password = req.body.password

  const user = await services.users.signUp(name, email, password, req.headers.host)

  return res.json(
    new SuccessResponse.Builder()
      .withContent(user)
      .build()
  )
}

exports.getUserCurrent = async (req, res) => {
  // Get token from request
  // const token = req.decoded.payload.email

  // const user = await services.users.findUserCurrent(token)

  // return res.json(
  //   new SuccessResponse.Builder()
  //     .withContent(user)
  //     .build()
  // )

  return res.json(
    new SuccessResponse.Builder()
      .withContent(req.userCurrent)
      .build()
  )
}
