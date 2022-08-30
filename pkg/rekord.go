package pkg

import "time"

type Importrekord struct {
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
type ImportHashedrekord struct {
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

// TLog holds current root hash and size of the merkle tree used to store the log entries.
type TLog interface {
	Size() (int64, error)
	Entry(index int64) (Entry, error)
}

type Exportrekord struct {
	apiVersion string
	Data       struct {
		Hash struct {
			Algorithm string `json:"algorithm"`
			Value     string `json:"value"`
		} `json:"hash"`
	} `json:"data"`
	Signature struct {
		Format    *string `json:"format,omitempty"`
		PublicKey *string `json:"publicKey,omitempty"`
		PGP       *string `json:"pgp,omitempty"`
		X509      *X509   `json:"x509,omitempty"`
	} `json:"signature"`
	Kind *string `json:"kind"`
}
type X509Extension struct {
	ID    *string `json:"id,omitempty"`
	Value *string `json:"value,omitempty"`
}

type X509 struct {
	Version            *int             `json:"version,omitempty"`
	SerialNumber       *string          `json:"serial_number,omitempty"`
	SignatureAlgorithm *string          `json:"signature_algorithm,omitempty"`
	IssuerOrganization *string          `json:"issuer_organization,omitempty"`
	IssuerCommonName   *string          `json:"issuer_common_name,omitempty"`
	ValidityNotBefore  *time.Time       `json:"validity_not_before,omitempty"`
	ValidityNotAfter   *time.Time       `json:"validity_not_after,omitempty"`
	Extensions         *[]X509Extension `json:"extensions,omitempty"`
}
type Exporthasedrekord struct {
	apiVersion string
	Data       struct {
		Hash struct {
			Algorithm string `json:"algorithm"`
			Value     string `json:"value"`
		} `json:"hash"`
	} `json:"data"`
	Signature struct {
		PublicKey *string `json:"publicKey,omitempty"`
		X509      *X509   `json:"x509,omitempty"`
	} `json:"signature"`
}

type ImportIntoto struct {
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
type tlog struct {
	treeID   string `json:"treeID"` //nolint:govet
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
	IntegratedTime int                `json:"integratedTime"`
	LogID          string             `json:"logID"`
	LogIndex       int                `json:"logIndex"`
	Kind           Kind               `json:"kind"`
	Rekord         *Exportrekord      `json:"rekord,omitempty"`
	HashedRekord   *Exporthasedrekord `json:"hashedrekord,omitempty"`
	Intoto         *Exporthasedrekord `json:"intoto,omitempty"`
	Raw            *string            `json:"raw,omitempty"`
}
type Kind struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}
