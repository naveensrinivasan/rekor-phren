package pkg

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/sigstore/rekor/pkg/generated/models"
	"golang.org/x/crypto/openpgp"
	"net/http"
	"strings"
)

const defaultHost = "https://rekor.sigstore.dev"

type tlog struct {
	treeID   string `json:"treeID"`
	TreeSize int64  `json:"treeSize"`
	host     string
}

// tlogEntry represents a single entry in a TLog.
type tlogEntry struct {
	Body           string       `json:"body"`
	IntegratedTime int          `json:"integratedTime"`
	LogID          string       `json:"logID"`
	LogIndex       int          `json:"logIndex"`
	Kind           string       `json:"kind"`
	APIVersion     string       `json:"apiVersion"`
	Rekord         Exportrekord `json:"rekord"`
}
type Entry struct {
	IntegratedTime int          `json:"integratedTime"`
	LogID          string       `json:"logID"`
	LogIndex       int          `json:"logIndex"`
	Kind           string       `json:"kind"`
	APIVersion     string       `json:"apiVersion"`
	Rekord         Exportrekord `json:"rekord"`
}
type Kind struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

type Rekor interface {
	models.Rekord | models.Alpine | models.Cose
}

func (t *tlog) Size() (int64, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/log", t.host), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var tlog tlog
	err = json.NewDecoder(resp.Body).Decode(&tlog)
	if err != nil {
		return 0, err
	}
	return tlog.TreeSize, nil
}

func (t tlog) Entry(index int) (Entry, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/log/entries?logIndex=%d", t.host, index))
	if err != nil {
		return Entry{}, err
	}
	defer resp.Body.Close()
	m := make(map[string]tlogEntry)
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return Entry{}, fmt.Errorf("error decoding response body: %v %s", err, resp.Status)
	}
	var val tlogEntry
	for _, v := range m {
		val = v
		break
	}
	value := getEntry(val)

	f, err := base64.StdEncoding.DecodeString(val.Body)
	if err != nil {
		return Entry{}, fmt.Errorf("error decoding base64: %v", err)
	}
	// Identify the kind of entry
	var k Kind
	err = json.Unmarshal(f, &k)
	if err != nil {
		return Entry{}, fmt.Errorf("error unmarshalling kind: %v", err)
	}
	value.Kind = k.Kind
	value.APIVersion = k.APIVersion
	switch value.Kind {
	case "rekord":
		rekord, err := handleRekord(f)
		if err != nil {
			return Entry{}, fmt.Errorf("error handling rekord: %v", err)
		}
		value.Rekord = rekord
	default:
		return Entry{}, fmt.Errorf("unsupported kind: %s", value.Kind)
	}
	return value, nil
}

func handleRekord(f []byte) (Exportrekord, error) {
	var i Importrekord
	err := json.Unmarshal(f, &i)
	if err != nil {
		return Exportrekord{}, fmt.Errorf("error unmarshalling importrekord: %v", err)
	}

	var e Exportrekord
	e.Kind = i.Kind
	e.apiVersion = i.ApiVersion
	e.Data.Hash.Algorithm = i.Spec.Data.Hash.Algorithm
	e.Data.Hash.Value = i.Spec.Data.Hash.Value
	e.Signature.Format = i.Spec.Signature.Format

	// convert the base64 encoded signature into a byte array for PublicKey.Content
	publicKey, err := base64.StdEncoding.DecodeString(i.Spec.Signature.PublicKey.Content)
	if err != nil {
		return Exportrekord{}, fmt.Errorf("error decoding public key: %v", err)
	}
	var pkeyText string
	if e.Signature.Format == "pgp" {
		pkeyText, err = getPublicKeyIdentities(string(publicKey))
	} else {
		pkeyText, err = getPublicKeyTextFromx509(string(publicKey))
	}
	if err != nil {
		return Exportrekord{}, fmt.Errorf("error getting public key identities: %v", err)
	}
	e.Signature.PublicKey = string(publicKey)
	// convert the base64 encoded signature into a byte array for Signature.Content
	signature, err := base64.StdEncoding.DecodeString(i.Spec.Signature.Content)
	if err != nil {
		return Exportrekord{}, fmt.Errorf("error decoding signature: %v", err)
	}
	e.Signature.Content = string(signature)
	e.Signature.PublicKeyText = pkeyText
	return e, nil
}

func (l tlog) Data(s []byte, kind string) (any, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	if kind == "" {
		return nil, fmt.Errorf("empty string for kind")
	}
	switch kind {
	case "rekord":
		var k models.RekordV001Schema
		// deserialize the data into a importrekord
		err := json.Unmarshal(s, &k)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling rekord: %v", err)
		}
		return k, nil
	case "alpine":
		var k models.HelmV001Schema
		// deserialize the data into a Alpine
		err := json.Unmarshal([]byte(s), &k)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling alpine: %v", err)
		}
		return k, nil
	}
	return nil, fmt.Errorf("unknown kind: %s", kind)
}

// TLog holds current root hash and size of the merkle tree used to store the log entries.
type TLog interface {
	Size() (int64, error)
	Entry(index int) (Entry, error)
	Data(data []byte, kind string) (any, error)
}

func NewTLog(host string) TLog {
	if host == "" {
		host = defaultHost
	}
	return &tlog{host: host}
}

//getPublicKeyIdentities returns the identities of the given public key.
func getPublicKeyIdentities(p string) (string, error) {
	b := strings.Builder{}
	result := ""
	r := bytes.NewReader([]byte(p))
	keys, err := openpgp.ReadArmoredKeyRing(r)
	if err != nil {

		return "", fmt.Errorf("Unable to read armored key ring: %s \n  %v\n", p, err)
	}
	for _, key := range keys {
		for name, _ := range key.Identities {
			b.WriteString(name)
			b.WriteString("\n")
		}
	}
	result = b.String()
	result = strings.ReplaceAll(result, "\n", "")
	return result, nil
}

//getEntry returns the entry from the given tlogEntry.
func getEntry(val tlogEntry) Entry {
	var value Entry
	value.LogID = val.LogID
	value.LogIndex = val.LogIndex
	value.Kind = val.Kind
	value.APIVersion = val.APIVersion
	value.IntegratedTime = val.IntegratedTime
	return value
}

func getPublicKeyTextFromx509(p string) (string, error) {
	certPEMBlock := []byte(p)
	pemBlock, _ := pem.Decode(certPEMBlock)
	publicKey, _ := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	publicKey.

	var blocks [][]byte
	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		fmt.Println(certDERBlock.Type)
		if certDERBlock.Type == "PUBLIC KEY" {
			blocks = append(blocks, certDERBlock.Bytes)
		}
	}
	var emails []string
	fmt.Println(len(blocks))
	for _, block := range blocks {
		cert, err := x509.ParsePKIXPublicKey(block)
		if err != nil {
			fmt.Println(err)
			continue
		}
		emails = append(emails, cert.EmailAddresses...)
		fmt.Println(emails)
	}
	return strings.Join(emails, "\n"), nil
}
