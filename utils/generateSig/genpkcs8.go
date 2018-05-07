package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

//https://stackoverflow.com/questions/48825863/how-to-create-pkcs8-private-key-using-go

type PKCS8Key struct {
	Version             int
	PrivateKeyAlgorithm []asn1.ObjectIdentifier
	PrivateKey          []byte
}

func MarshalPKCS8PrivateKey(key *rsa.PrivateKey) ([]byte, error) {
	var pkey PKCS8Key
	pkey.Version = 0
	pkey.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 1)
	pkey.PrivateKeyAlgorithm[0] = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	pkey.PrivateKey = x509.MarshalPKCS1PrivateKey(key)
	return asn1.Marshal(pkey)
}

func main() {
	// Generate the private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	fatal(err)

	// Encode the private key into PEM data.
	bytes, err := MarshalPKCS8PrivateKey(privateKey)
	fatal(err)
	privatePem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: bytes,
		},
	)

	ioutil.WriteFile("test0.pem", privatePem, 0440)

	publicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)

	fatal(err)

	publicKeyEncoded := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKey,
		},
	)

	ioutil.WriteFile("test0.pub", publicKeyEncoded, 0440)
	fmt.Printf("%s\n", privatePem)
	// -----BEGIN PRIVATE KEY-----
	// MIIEvAIBADALBgkqhkiG9w0BAQEEggSoMIIEpAIBAAKCAQEAz5xD5cdqdE0PMmk1
	// 4YN6Tj0ybTsvS5C95ogQmBJ4bGxiuGPR5JtIc+UmT8bnCHtK5xnHiP+gPWunwmhS
	// ...
	// qpb1NZsMLz2lRXqx+3Pq7Q==
	// -----END PRIVATE KEY-----
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
