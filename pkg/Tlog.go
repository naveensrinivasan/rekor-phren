package pkg

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	//nolint
	"golang.org/x/crypto/openpgp"
)

const defaultHost = "https://rekor.sigstore.dev"

// NewTLog creates an instance of the Tlog.
func NewTLog(host string) TLog {
	if host == "" {
		host = defaultHost
	}
	return &tlog{host: host}
}

// Size returns the size of the last entry.
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
	//nolint
	defer resp.Body.Close()
	var log tlog
	err = json.NewDecoder(resp.Body).Decode(&log)
	if err != nil {
		return 0, err
	}
	return log.TreeSize, nil
}

// Entry returns the entry from the given tlogEntry.
func (t *tlog) Entry(index int64) (Entry, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/log/entries?logIndex=%d", t.host, index))
	if err != nil {
		return Entry{}, err
	}
	//nolint
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
	var k Kind
	err = json.Unmarshal(f, &k)
	if err != nil {
		return Entry{}, fmt.Errorf("error unmarshalling kind: %v", err)
	}
	value.Kind = Kind{
		APIVersion: k.APIVersion,
		Kind:       k.Kind,
	}
	switch value.Kind.Kind {
	case "rekord":
		rekord, err := handleRekord(f)
		if err != nil {
			return Entry{}, fmt.Errorf("error handling rekord: %v", err)
		}
		value.Rekord = &rekord
	case "hashedrekord":
		rekord, err := handleHashedRekord(f)
		if err != nil {
			return Entry{}, fmt.Errorf("error handling hashedrekord: %v", err)
		}
		value.HashedRekord = &rekord
	case "intoto":
		intoto, err := handleIntoto(f)
		if err != nil {
			return Entry{}, fmt.Errorf("error handling intoto: %w", err)
		}
		value.Intoto = &intoto
	default:
		return value, nil
	}
	value.Date = time.Now()
	return value, nil
}

// handleIntoto handles the intoto entry.
func handleIntoto(s []byte) (InToTo, error) {
	var i importIntoto
	err := json.Unmarshal(s, &i)
	if err != nil {
		return InToTo{}, fmt.Errorf("error unmarshalling importrekord: %w", err)
	}
	var e InToTo
	e.apiVersion = i.APIVersion
	e.Data.Hash.Algorithm = i.Spec.Content.Hash.Algorithm
	e.Data.Hash.Value = i.Spec.Content.Hash.Value

	// convert the base64 encoded signature into a byte array for PublicKey.Content
	publicKey, err := base64.StdEncoding.DecodeString(i.Spec.PublicKey)
	if err != nil {
		return InToTo{}, fmt.Errorf("error decoding public key: %w", err)
	}
	p := string(publicKey)
	e.Signature.PublicKey = p
	identity, err := getx509Identity(string(publicKey))
	if err == nil {
		e.Signature.X509 = identity
	}

	return e, nil
}

// handleHashRekord handles the hashedrekord entry.
func handleHashedRekord(s []byte) (Hashedrekord, error) {
	var i importHashedrekord
	err := json.Unmarshal(s, &i)
	if err != nil {
		return Hashedrekord{}, fmt.Errorf("error unmarshalling importrekord: %w", err)
	}
	var e Hashedrekord
	e.apiVersion = i.APIVersion
	e.Data.Hash.Algorithm = i.Spec.Data.Hash.Algorithm
	e.Data.Hash.Value = i.Spec.Data.Hash.Value

	// convert the base64 encoded signature into a byte array for PublicKey.Content
	publicKey, err := base64.StdEncoding.DecodeString(i.Spec.Signature.PublicKey.Content)
	if err != nil {
		return Hashedrekord{}, fmt.Errorf("error decoding public key: %w", err)
	}
	p := string(publicKey)
	e.Signature.PublicKey = p
	identity, err := getx509Identity(string(publicKey))
	if err == nil {
		e.Signature.X509 = identity
	}

	return e, nil
}

// handleRekord handles the rekord entry
func handleRekord(f []byte) (Rekord, error) {
	var i importrekord
	err := json.Unmarshal(f, &i)
	if err != nil {
		return Rekord{}, fmt.Errorf("error unmarshalling importrekord: %v", err)
	}

	var e Rekord
	e.apiVersion = i.APIVersion
	e.Data.Hash.Algorithm = i.Spec.Data.Hash.Algorithm
	e.Data.Hash.Value = i.Spec.Data.Hash.Value

	e.Signature.Format = i.Spec.Signature.Format

	// convert the base64 encoded signature into a byte array for PublicKey.Content
	publicKey, err := base64.StdEncoding.DecodeString(i.Spec.Signature.PublicKey.Content)
	if err != nil {
		return Rekord{}, fmt.Errorf("error decoding public key: %v", err)
	}
	var pkeyText string
	p := string(publicKey)
	format := i.Spec.Signature.Format
	e.Signature.PublicKey = p
	if format == "pgp" {
		pkeyText, err = getPGPIdentity(string(publicKey))
		if err != nil {
			return Rekord{}, fmt.Errorf("error getting public key identities: %v", err)
		}
		e.Signature.PGP = pkeyText
	} else {
		identity, err := getx509Identity(string(publicKey))
		if err == nil {
			e.Signature.X509 = identity
		}
	}
	return e, nil
}

// getPGPIdentity returns the identities of the given public key.
func getPGPIdentity(p string) (string, error) {
	b := strings.Builder{}
	result := ""
	r := bytes.NewReader([]byte(p))
	keys, err := openpgp.ReadArmoredKeyRing(r)
	if err != nil {
		return "", fmt.Errorf("Unable to read armored key ring: %s \n  %v\n", p, err)
	}
	for _, key := range keys {
		for name := range key.Identities {
			b.WriteString(name)
			b.WriteString("\n")
		}
	}
	result = b.String()
	result = strings.ReplaceAll(result, "\n", "")
	return result, nil
}

// getEntry returns the entry from the given tlogEntry.
func getEntry(val tlogEntry) Entry {
	var value Entry
	value.LogID = val.LogID
	value.LogIndex = val.LogIndex
	value.Kind = Kind{
		Kind:       val.Kind,
		APIVersion: val.APIVersion,
	}
	value.IntegratedTime = val.IntegratedTime
	return value
}

// getx509Identity returns the identities of the given public key.
func getx509Identity(publicKey string) (*X509, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	serialNumber := cert.SerialNumber.String()
	signatureAlgorithm := cert.SignatureAlgorithm.String()
	var extension []X509Extension //nolint:prealloc
	validCerts := map[string]bool{
		"2.5.29.17":             true,
		"1.3.6.1.4.1.57264.1.1": true,
		"1.3.6.1.4.1.57264.1.2": true,
		"1.3.6.1.4.1.57264.1.3": true,
		"1.3.6.1.4.1.57264.1.4": true,
		"1.3.6.1.4.1.57264.1.5": true,
		"1.3.6.1.4.1.57264.1.6": true,
	}

	for _, e := range cert.Extensions {
		if _, ok := validCerts[e.Id.String()]; !ok {
			continue
		}
		id := e.Id.String()
		value := string(e.Value)
		extension = append(extension, X509Extension{
			ID:    id,
			Value: value,
		})
	}
	certificate := X509{
		Version:            cert.Version,
		SerialNumber:       serialNumber,
		SignatureAlgorithm: signatureAlgorithm,
		IssuerOrganization: cert.Issuer.Organization[0],
		IssuerCommonName:   cert.Issuer.CommonName,
		ValidityNotBefore:  &cert.NotBefore,
		ValidityNotAfter:   &cert.NotAfter,
		Extensions:         extension,
	}
	return &certificate, nil
}
