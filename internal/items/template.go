package items

import (
	"bytes"
	"encoding/json"
	"text/template"
)

func RenderTemplate(template []byte, params []byte) ([]byte, error) {
	tpl, err := ValidateTemplate(template)
	if err != nil {
		return nil, err
	}

	p := make(map[string]any)
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	var res bytes.Buffer
	if err := tpl.Execute(&res, p); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func ValidateTemplate(tpl []byte) (*template.Template, error) {
	return template.New("item").Parse(string(tpl))
}
