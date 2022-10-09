package pkg

import "time"

type importIntoto struct {
	APIVersion string `json:"apiVersion"`
	Spec       struct {
		Content struct {
			Hash struct {
				Algorithm string `json:"algorithm"`
				Value     string `json:"value"`
			} `json:"hash"`
		} `json:"content"`
		PublicKey string `json:"publicKey"`
	} `json:"spec"`
	Kind string `json:"kind"`
}

// tlogEntry represents a single entry in a TLog.
type tlogEntry struct {
	Body           string `json:"body"`
	IntegratedTime int    `json:"integratedTime"`
	LogID          string `json:"logID"`
	LogIndex       int    `json:"logIndex"`
	Kind           string `json:"kind"`
	APIVersion     string `json:"apiVersion"`
}
type importrekord struct {
	APIVersion string `json:"apiVersion"`
	Spec       struct {
		Data struct {
			Hash struct {
				Algorithm string `json:"algorithm"`
				Value     string `json:"value"`
			} `json:"hash"`
		} `json:"data"`
		Signature struct {
			Content   string `json:"content"`
			Format    string `json:"format"`
			PublicKey struct {
				Content string `json:"content"`
			} `json:"publicKey"`
		} `json:"signature"`
	} `json:"spec"`
	Kind string `json:"kind"`
}
type importHashedrekord struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		Data struct {
			Hash struct {
				Algorithm string `json:"algorithm"`
				Value     string `json:"value"`
			} `json:"hash"`
		} `json:"data"`
		Signature struct {
			Content   string `json:"content"`
			PublicKey struct {
				Content string `json:"content"`
			} `json:"publicKey"`
		} `json:"signature"`
	} `json:"spec"`
}

type tlog struct {
	TreeSize int `json:"treeSize"`
	host     string
}

// TLog holds current root hash and size of the merkle tree used to store the log entries.
type TLog interface {
	Size() (int, error)
	Entry(index int) (Entry, error)
}
type RekordData struct {
	Hash RekorDataHash `json:"hash"`
}
type RekorDataHash struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}
type Signature struct {
	Format    string `json:"format,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
	PGP       string `json:"pgp,omitempty"`
	X509      *X509  `json:"x509,omitempty"`
}
type Rekord struct {
	apiVersion string
	Data       RekordData `json:"data"`
	Signature  Signature  `json:"signature"`
}
type X509Extension struct {
	ID    string `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
}

type X509 struct {
	Version            int             `json:"version,omitempty"`
	SerialNumber       string          `json:"serial_number,omitempty"`
	SignatureAlgorithm string          `json:"signature_algorithm,omitempty"`
	IssuerOrganization string          `json:"issuer_organization,omitempty"`
	IssuerCommonName   string          `json:"issuer_common_name,omitempty"`
	ValidityNotBefore  time.Time       `json:"validity_not_before,omitempty"`
	ValidityNotAfter   time.Time       `json:"validity_not_after,omitempty"`
	Extensions         []X509Extension `json:"extensions,omitempty"`
}
type Hashedrekord struct {
	apiVersion string
	Data       RekordData      `json:"data"`
	Signature  RekordSignature `json:"signature"`
}
type RekordSignature struct {
	PublicKey string `json:"publicKey,omitempty"`
	X509      *X509  `json:"x509,omitempty"`
}
type InToTo struct {
	apiVersion string
	Data       RekordData      `json:"data"`
	Signature  RekordSignature `json:"signature"`
}
type Entry struct {
	IntegratedTime int           `json:"integratedTime"`
	LogID          string        `json:"logID"`
	LogIndex       int           `json:"logIndex"`
	Kind           Kind          `json:"kind"`
	Rekord         *Rekord       `json:"rekord,omitempty"`
	HashedRekord   *Hashedrekord `json:"hashedrekord,omitempty"`
	Intoto         *InToTo       `json:"intoto,omitempty"`
	Date           time.Time     `json:"date"`
}
type Kind struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}
