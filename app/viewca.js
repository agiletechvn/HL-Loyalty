const x509 = require("x509");
var fs = require("fs");
var path = require('path');
var os = require('os');

const user = process.argv[2] || 'user1'
var store_path = path.join(os.homedir(), '.hfc-key-store');
console.log(' Store path:'+store_path);
var obj = JSON.parse(
  fs.readFileSync(store_path + '/' + user
  , "utf8")
);
var cert = x509.parseCert(obj.enrollment.identity.certificate);
console.log(JSON.stringify(cert, null, 2));
