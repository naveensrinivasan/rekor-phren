package pkg

type Importrekord struct {
	ApiVersion string `json:"apiVersion"`
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
type Exportrekord struct {
	apiVersion string
	Data       struct {
		Hash struct {
			Algorithm string `json:"algorithm"`
			Value     string `json:"value"`
		} `json:"hash"`
	} `json:"data"`
	Signature struct {
		Content       string `json:"content"`
		Format        string `json:"format"`
		PublicKey     string `json:"publicKey"`
		PublicKeyText string `json:"publicKeyText"`
	} `json:"signature"`
	Kind string `json:"kind"`
}
