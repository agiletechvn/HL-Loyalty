# create enrollment
CERT_FILE=$1
PRIVATE_KEY=$2
CERTIFICATE=$(cat $CERT_FILE | sed 's/$/\\r\\n/' | tr -d '\n')
PRIVATE_KEY_NAME=`basename $PRIVATE_KEY | sed 's/_sk//'`
NAME=`basename $CERT_FILE | sed 's/\.pem//'`
MSPID=Org1MSP

cat << EOF > ./hfc-key-store/$NAME
{
  "name": "$NAME",
  "mspid": "$MSPID",
  "roles": null,
  "affiliation": "",
  "enrollmentSecret": "",
  "enrollment": {
    "signingIdentity": "$PRIVATE_KEY_NAME",
    "identity": {
      "certificate": "$CERTIFICATE"
    }
  }
}
EOF

cp $PRIVATE_KEY ./hfc-key-store/${PRIVATE_KEY_NAME}-priv

echo "created $NAME successfully ..."  
