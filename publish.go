package main

//import "github.com/giorgisio/goav/avformat"
import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// https://stackoverflow.com/questions/33450980/golang-remove-all-contents-of-a-directory
func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func EncryptFile(inputFilename string, outputFilename string) (error, []byte) {
	plaintext, err := ioutil.ReadFile(inputFilename)

	if err != nil {
		panic(err.Error())
	}

	// AES-256
	key := make([]byte, 32)

	_, er := rand.Read(key)

	if er != nil {
		return er, nil
	}

	block, err := aes.NewCipher(key)

	if err != nil {
		return err, nil
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	f, err := os.Create(outputFilename)

	if err != nil {
		return err, nil
	}

	_, err = io.Copy(f, bytes.NewReader(ciphertext))

	if err != nil {
		return err, nil
	}

	return nil, key
}

//https://gist.github.com/josephspurrier/12cc5ed76d2228a41ceb
func DecryptFile(key []byte, inputpath string) (error, []byte) {

	ciphertext, err := ioutil.ReadFile(inputpath)

	if err != nil {
		return err, nil
	}
	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err, nil
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		return errors.New("Text is too short"), nil
	}

	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return nil, ciphertext

}

// https://stackoverflow.com/questions/20655702/signing-and-decoding-with-rsa-sha-in-go
// loadPrivateKey loads an parses a PEM encoded private key file.
func loadPublicKey(path string) (Unsigner, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}
	return parsePublicKey(data)
}

// parsePublicKey parses a PEM encoded private key.
func parsePublicKey(pemBytes []byte) (Unsigner, error) {
	block, _ := pem.Decode(pemBytes)

	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "PUBLIC KEY":
		rsa, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}

	return newUnsignerFromKey(rawkey)
}

// loadPrivateKey loads an parses a PEM encoded private key file.
func loadPrivateKey(path string) (Signer, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivateKey(data)
}

// parsePublicKey parses a PEM encoded private key.
func parsePrivateKey(pemBytes []byte) (Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	fmt.Println(block.Type)
	switch block.Type {
	case "PRIVATE KEY":
		rsa, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}
	return newSignerFromKey(rawkey)
}

// A Signer is can create signatures that verify against a public key.
type Signer interface {
	// Sign returns raw signature for the given data. This method
	// will apply the hash specified for the keytype to the data.
	Sign(data []byte) ([]byte, error)
}

// A Signer is can create signatures that verify against a public key.
type Unsigner interface {
	// Sign returns raw signature for the given data. This method
	// will apply the hash specified for the keytype to the data.
	Unsign(message []byte, sig []byte) error
}

func newSignerFromKey(k interface{}) (Signer, error) {
	var sshKey Signer
	switch t := k.(type) {
	case *rsa.PrivateKey:
		sshKey = &rsaPrivateKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

func newUnsignerFromKey(k interface{}) (Unsigner, error) {
	var sshKey Unsigner
	switch t := k.(type) {
	case *rsa.PublicKey:
		sshKey = &rsaPublicKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

type rsaPublicKey struct {
	*rsa.PublicKey
}

type rsaPrivateKey struct {
	*rsa.PrivateKey
}

// Sign signs data with rsa-sha256
func (r *rsaPrivateKey) Sign(data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	d := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.SHA256, d)
}

// Unsign verifies the message using a rsa-sha256 signature
func (r *rsaPublicKey) Unsign(message []byte, sig []byte) error {
	h := sha256.New()
	h.Write(message)
	d := h.Sum(nil)
	return rsa.VerifyPKCS1v15(r.PublicKey, crypto.SHA256, d, sig)
}

// call with publish.go American_Idiot.mp3
func main() {

	segmentLength := "15"
	if len(os.Args) < 2 {
		panic("Not enough arguments")
	}
	filename := os.Args[1]
	chunkPath := "chunks/"
	encChunkDir := "encChunks/"
	decChunkDir := "decChunks/"

	_ = os.Mkdir(chunkPath, os.ModePerm)
	RemoveContents(chunkPath)

	_ = os.Mkdir(encChunkDir, os.ModePerm)
	RemoveContents(encChunkDir)

	_ = os.Mkdir(decChunkDir, os.ModePerm)
	RemoveContents(decChunkDir)

	fmt.Println("Using file " + filename)
	cmd := exec.Command("ffmpeg", "-i", filename, "-f", "segment",
		"-segment_time", segmentLength, "-c", "copy", chunkPath+"chunk%03d.mp3")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error while executing ffmpeg ", err)
	}

	files, err := ioutil.ReadDir("chunks")
	if err != nil {
		panic(err)
	}

	signer, err := loadPrivateKey("test0.pem")
	if err != nil || signer == nil {
		fmt.Errorf("signer is damaged: %v", err)
	}

	fmt.Println(signer)

	// Encrypt then sign
	for _, file := range files {
		// TODO: create a struct for a song to publish to the blockchain
		fmt.Println("Encrypting " + file.Name())
		inputpath := chunkPath + file.Name()
		encChunkPath := encChunkDir + file.Name() + ".enc"
		encChunkSigPath := encChunkDir + file.Name() + ".sig"
		encChunkKeyPath := encChunkDir + file.Name() + ".key"

		err, key := EncryptFile(inputpath, encChunkPath)
		ioutil.WriteFile(encChunkKeyPath, key, 0440)

		ciphertext, err := ioutil.ReadFile(encChunkPath)

		//fmt.Println("size of ciphertext ", len(ciphertext))

		if err != nil {
			panic(err)
		}

		signed, err := signer.Sign(ciphertext)

		if err != nil {
			fmt.Errorf("could not sign request: %v", err)
		}

		//sig := base64.StdEncoding.EncodeToString(signed)
		//fmt.Printf("Signature: %v\n", sig)
		ioutil.WriteFile(encChunkSigPath, signed, 0440)

		/* test sig verification
		parser, err := loadPublicKey("test0.pub")
		if err != nil {
			fmt.Errorf("could not load public key: %v", err)
		}

		perr := parser.Unsign(ciphertext, signed)
		if perr != nil {
			fmt.Errorf("could not unsign request: %v", err)
		}
		fmt.Println("Signature verified")

		*/

		/* test decrypting
		   err, plaintext := DecryptFile(key, encChunkPath)

		   fmt.Println(err, len(plaintext))

		   outFilePath := decChunkDir + file.Name()
		   ioutil.WriteFile(outFilePath, plaintext, 0644)
		*/

	}

	RemoveContents(chunkPath)

}
