package items

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Item
// +kubebuilder:pruning:PreserveUnknownFields
type Item unstructured.Unstructured

func (i *Item) DeepCopyInto(i2 *Item) {
	i2.Object = runtime.DeepCopyJSON(i.Object)
}

// ParamSchema
// +kubebuilder:pruning:PreserveUnknownFields
type ParamSchema unstructured.Unstructured

func (ps ParamSchema) Empty() bool {
	return len(ps.Object) == 0
}

func (ps ParamSchema) DeepCopyInto(ps2 *ParamSchema) {
	ps2.Object = runtime.DeepCopyJSON(ps.Object)
}

// Params
// +kubebuilder:pruning:PreserveUnknownFields
type Params unstructured.Unstructured

func (p *Params) DeepCopyInto(p2 *Params) {
	p2.Object = runtime.DeepCopyJSON(p.Object)
}

func (p *Params) DeepCopy() *Params {
	if p == nil {
		return nil
	}
	out := new(Params)
	p.DeepCopyInto(out)
	return out
}

// TemplateItem
// +kubebuilder:object:generate=true
type TemplateItem struct {
	// +kubebuilder:validation:Required
	Item        Item        `json:"item"`
	ParamSchema ParamSchema `json:"paramSchema,omitempty"`
}

type (
	TemplateItems map[string]TemplateItem
	Items         map[string]*unstructured.Unstructured
)

// TemplateParams
// +kubebuilder:object:generate=true
type TemplateParams map[string]Params
