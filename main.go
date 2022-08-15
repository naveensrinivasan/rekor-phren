package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/naveensrinivasan/rekor-phren/pkg"
)

func main() {
	k := pkg.NewTLog("")
	data, err := k.Entry(2001)
	if err != nil {
		fmt.Println(err)
	}
	serialized, err := Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(serialized))

	//main1()
}

// Marshal is a UTF-8 friendly marshaller.  Go's json.Marshal is not UTF-8
// friendly because it replaces the valid UTF-8 and JSON characters "&". "<",
// ">" with the "slash u" unicode escaped forms (e.g. \u0026).  It preemptively
// escapes for HTML friendliness.  Where text may include any of these
// characters, json.Marshal should not be used. Playground of Go breaking a
// title: https://play.golang.org/p/o2hiX0c62oN
func Marshal(i interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(i)
	return bytes.TrimRight(buffer.Bytes(), "\n"), err
}

/*
func main1() {
	i, err := getPublicKeyIdentities(pubKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(i)
}

const pubKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
pa0EbEfq
=dMc+
-----END PGP PUBLIC KEY BLOCK-----`
*/
