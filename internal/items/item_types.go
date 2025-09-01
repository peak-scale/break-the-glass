package items

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:generate=true

// TemplateItem
type TemplateItem struct {
	// +kubebuilder:validation:Required
	ManifestTemplate runtime.RawExtension `json:"manifestTemplate"`
	ParamSchema      runtime.RawExtension `json:"paramSchema,omitempty"`
}

type (
	// TemplateItems
	TemplateItems map[string]TemplateItem

	// Items
	Items map[string]*runtime.RawExtension
)

// TemplateParams
type TemplateParams map[string]runtime.RawExtension
