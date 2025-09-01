package items

import (
	"bytes"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func RenderTemplate(i Item, params Params) (*unstructured.Unstructured, error) {
	tpl, err := ValidateTemplate(i)
	if err != nil {
		return nil, err
	}
	var res bytes.Buffer
	if err := tpl.Execute(&res, params.Object); err != nil {
		return nil, err
	}
	out := &unstructured.Unstructured{}
	err = yaml.Unmarshal(res.Bytes(), &out.Object)
	return out, err
}

func ValidateTemplate(i Item) (*template.Template, error) {
	it, err := yaml.Marshal(i.Object)
	if err != nil {
		return nil, err
	}
	return template.New("item").Parse(string(it))
}
