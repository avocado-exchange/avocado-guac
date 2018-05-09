package main

//import "github.com/giorgisio/goav/avformat"
import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
	"github.com/dkaps125/go-contract/contract"
	mp3 "github.com/hajimehoshi/go-mp3"
)

const catalogAddress string = "0x8f0483125fcb9aaaefa9209d8e9d7b9c8b9fb90f"

var c contract.Contract

type FileData struct {
	Chunks map[string][]byte
}

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

	return Decrypt(key, ciphertext)
}

func Decrypt(key []byte, ciphertext []byte) (error, []byte) {

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

// ffmpeg util functions

func chunkFile(filename string, chunkPath string, segmentLength string) ([]os.FileInfo, error) {
	cmd := exec.Command("ffmpeg", "-i", filename, "-f", "segment",
		"-segment_time", segmentLength, "-c", "copy", chunkPath+"chunk%03d-"+filename)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error while executing ffmpeg ", err)
		return nil, err
	}

	files, err := ioutil.ReadDir(chunkPath)
	if err != nil {
		return nil, err
	}

	return files, err
}

func unChunkFile(outputPath string, chunkPath string) error {
	files, err := ioutil.ReadDir(chunkPath)
	if err != nil {
		return err
	}

	ffmpegInput := "concat:"
	for _, file := range files[:len(files)-1] {
		ffmpegInput += chunkPath + file.Name() + "|"
	}

	ffmpegInput += chunkPath + files[len(files)-1].Name()
	fmt.Println(ffmpegInput)

	cmd := exec.Command("ffmpeg", "-i", ffmpegInput, "-acodec",
		"copy", outputPath)

	return cmd.Run()
}

func getMetadata(filepath string) (tag.Metadata, int64, error) {
	mp3File, err := os.Open(filepath)
	defer mp3File.Close()

	if err != nil {
		return nil, -1, err
	}

	d, err := mp3.NewDecoder(mp3File)
	if err != nil {
		return nil, -1, err
	}

	defer d.Close()

	mp3File.Seek(0, 0)

	meta, err := tag.ReadFrom(mp3File)
	if err != nil {
		return nil, -1, err
	}

	return meta, d.Length(), nil
}

func loadCatalogContract() (contract.Contract, error) {
	var c contract.Contract
	c, err := c.Init("../guac-client/src/contracts/Catalog.json", catalogAddress, "http://localhost:9545")

	return c, err
}

func encryptAndChunk(filename string, cost uint32, myAccount string) ([][32]byte, *FileData) {
	var hashes [][32]byte
	var chunkStruct FileData

	chunkStruct.Chunks = make(map[string][]byte)

	segmentLength := "15"

	chunkPath := "chunks/"
	encChunkDir := "encChunks/"
	decChunkDir := "decChunks/"

	meta, songLen, err := getMetadata(filename)

	if err == nil {
		fmt.Println("Parsed metadata for: " + meta.Title())
	} else {
		panic(err)
	}

	_ = os.Mkdir(chunkPath, os.ModePerm)
	RemoveContents(chunkPath)

	_ = os.Mkdir(encChunkDir, os.ModePerm)
	RemoveContents(encChunkDir)

	_ = os.Mkdir(decChunkDir, os.ModePerm)
	RemoveContents(decChunkDir)

	fmt.Println("Using file " + filename)

	files, err := chunkFile(filename, chunkPath, segmentLength)

	signer, err := loadPrivateKey("test0.pem")
	if err != nil || signer == nil {
		_ = fmt.Errorf("signer is damaged: %v", err)
		panic(err)
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

		if err != nil {
			panic(err)
		}

		ciphertext, err := ioutil.ReadFile(encChunkPath)
		cHash := sha256.Sum256(ciphertext)
		hashes = append(hashes, cHash)
		chunkStruct.Chunks[hex.EncodeToString(cHash[:])] = ciphertext

		//fmt.Println("size of ciphertext ", len(ciphertext))

		if err != nil {
			panic(err)
		}

		signed, err := signer.Sign(ciphertext)

		if err != nil {
			_ = fmt.Errorf("could not sign request")
			panic(err)
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

		/* decryption example
		err, plaintext := DecryptFile(key, encChunkPath)

		fmt.Println(err, len(plaintext))

		outFilePath := decChunkDir + file.Name()
		ioutil.WriteFile(outFilePath, plaintext, 0644)
		*/

	}

	/* post-decryption unchunk example
	processedFileName := "processed_" + filename
	os.Remove(processedFileName)
	err = unChunkFile(processedFileName, decChunkDir)

	if err != nil {
		panic(err)
	}

	*/
	RemoveContents(chunkPath)

	/* List song on the blockchain */
	// TODO: parse filename properly

	var (
		filenameBytes [32]byte
		title         [32]byte
		artist        [32]byte
		album         [32]byte
		genre         [32]byte
		filetype      uint32
	)

	filetype = 0
	copy(filenameBytes[:], filename)
	copy(title[:], meta.Title())
	copy(artist[:], meta.Artist())
	copy(album[:], meta.Album())
	copy(genre[:], meta.Genre())

	// this is obnoxious

	s, err := c.Transact("listSong", myAccount, cost, filetype,
		filenameBytes, title, artist, album,
		genre, uint32(meta.Year()), uint32(songLen), uint32(len(files)))

	fmt.Println(s)
	if err != nil {
		panic(err)
	}

	return hashes, &chunkStruct
}

func publishChunks(hashes [][32]byte, myAccount string, filename string, chData *FileData) func([]interface{}) error {
	return func(data []interface{}) error {
		songNum := data[1].(uint32)
		fmt.Printf("Publishing song %d\n", songNum)

		var host [32]byte
		copy(host[:], "http://localhost:6548")

		c.Transact("publishChunks", myAccount, songNum, hashes, host)
		eventNum, _ := c.RegisterEventListener("RandomnessReady")
		c.ListenOnce(eventNum, "RandomnessReady", revealChunks(myAccount, songNum, filename, chData))

		return nil
	}
}

func bytesToHostname(bytes [][32]byte) []string {
	var res []string

	for _, v := range bytes {
		res = append(res, string(v[:]))
	}

	return res
}

func revealChunks(myAccount string, songNum uint32, filename string, chData *FileData) func([]interface{}) error {
	return func(data []interface{}) error {
		chunkInd1 := data[0].(*big.Int)
		chunkInd2 := data[1].(*big.Int)
		csAddresses := data[2].([][32]byte)
		key1, err1 := ioutil.ReadFile("encChunks/chunk" + fmt.Sprintf("%03d", chunkInd1) + "-" + filename + ".key")
		key2, err2 := ioutil.ReadFile("encChunks/chunk" + fmt.Sprintf("%03d", chunkInd2) + "-" + filename + ".key")

		if err1 != nil {
			panic(err1)
		}

		if err2 != nil {
			panic(err2)
		}

		var k1 [32]byte
		var k2 [32]byte

		copy(k1[:], key1)
		copy(k2[:], key2)

		c.Transact("revealChunks", myAccount, k1, k2, songNum)

		sendChunks(bytesToHostname(csAddresses), chData)
		return nil
	}
}

func sendChunks(hostnames []string, chData *FileData) {
	jsonStr, _ := json.Marshal(chData)
	host := strings.Trim(hostnames[0], " \n\000")
	fmt.Printf("sending to %s\n", host)
	req, err := http.NewRequest("PUT", host, bytes.NewBuffer(jsonStr))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	//body, _ := ioutil.ReadAll(resp.Body)
}

func getChunk(hostnames []string, chunkHash [32]byte) []byte {
	host := strings.Trim(hostnames[0], " \n\000")
	url := host + "/?hash=" + hex.EncodeToString(chunkHash[:])

	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func getChunks(hostnames []string, chunkHashes [][32]byte) [][]byte {
	res := make([][]byte, 1)

	for _, v := range chunkHashes {
		res = append(res, getChunk(hostnames, v))
	}

	return res
}

func preview(song uint32, outputFilePath string) {
	val, _ := c.Call("getPreview", song)
	chunk1Hash := val[0].([32]byte)
	chunk2Hash := val[1].([32]byte)
	chunk1Key := val[2].([32]byte)
	chunk2Key := val[3].([32]byte)
	csAddresses := val[4].([][32]byte)

	fmt.Printf("%v %v %v %v\n", chunk1Hash, chunk2Hash, chunk1Key, chunk2Key)
	chunk1 := getChunk(bytesToHostname(csAddresses), chunk1Hash)
	//chunk2 := getChunk(bytesToHostname(csAddresses), chunk2Hash)

	fmt.Printf("%v\n", chunk1)
	//fmt.Printf("%v\n", chunk2)

	err, plaintext := Decrypt(chunk1Key[:], chunk1)

	fmt.Println(err, len(plaintext))

	ioutil.WriteFile(outputFilePath, plaintext, 0644)
}

func purchase(song uint32, myAccount string, value int64, savePath string) {
	c.TransactValue("purchaseSong", myAccount, value, song)
	eventNum, _ := c.RegisterEventListener("ListingPurchased")
	c.ListenOnce(eventNum, "ListingPurchased", handlePurchase(song, savePath))
}

func handlePurchase(songNum uint32, savePath string) func([]interface{}) error {
	return func(data []interface{}) error {
		//sellerBytes := data[2].([32]byte)
		//seller := string(sellerBytes[:])
		hashes := data[3].([][32]byte)
		cs := data[4].([][32]byte)

		//keys := getKeys(seller, hashes)

		chunks := getChunks(bytesToHostname(cs), hashes)

		fmt.Printf("%v\n", chunks)

		/*
			decChunkDir := "decChunks/"
			_ = os.Mkdir(decChunkDir, os.ModePerm)
			RemoveContents(decChunkDir)

			for i, chunk := range chunks {
				err, plaintext := Decrypt(chunk, keys[i][:])

				fmt.Println(err, len(plaintext))

				ioutil.WriteFile(decChunkDir+"_chunk"+i+".mp3", plaintext, 0644)
			}

			/* post-decryption unchunk example

			err = unChunkFile(savePath, decChunkDir)

			if err != nil {
				panic(err)
			}
		*/

		return nil
	}
}

func getKeys(hostname string, hashes [][32]byte) [][32]byte {
	return make([][32]byte, 1)
}

func serveKeys() {

}

// call with
// go run publish.go publish chop_suey.mp3 100 0x627306090abab3a6e1400e9345bc60c78a8bef57
func main() {
	var err error

	fmt.Println("Loading contract...")
	c, err = loadCatalogContract()

	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		panic("Not enough args. Expected command type")
	}

	if os.Args[1] == "preview" {
		if len(os.Args) < 4 {
			panic("Missing song number")
		}

		songNum, _ := strconv.Atoi(os.Args[2])
		preview(uint32(songNum), os.Args[3])
		return
	}

	if os.Args[1] == "purchase" {
		if len(os.Args) < 6 {
			panic("Missing song number, account, or output file path")
		}

		songNum, _ := strconv.Atoi(os.Args[2])
		val, _ := strconv.Atoi(os.Args[4])
		myAccount := os.Args[3]
		outPath := os.Args[4]
		purchase(uint32(songNum), myAccount, int64(val), outPath)
		return
	}

	if os.Args[1] != "publish" {
		panic("Invalid command - should be one of publish, preview, or purchase")
	}

	//go serveKeys()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("# ")

		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, " \n")
		toks := strings.Split(text, " ")

		if len(toks) < 3 {
			panic("Not enough arguments")
		}

		filename := toks[0]
		cost, err := strconv.Atoi(toks[1])
		if err != nil {
			panic(err)
		}
		myAddress := toks[2]
		hashes, chunkData := encryptAndChunk(filename, uint32(cost), myAddress)
		eventNum, _ := c.RegisterEventListener("SongListed")
		c.ListenOnce(eventNum, "SongListed", publishChunks(hashes, myAddress, filename, chunkData))

	}
}
