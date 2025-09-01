package controller

import "k8s.io/apimachinery/pkg/runtime"

var (
	mtConfigMapParameterized = runtime.RawExtension{Raw: []byte(`
{
  "kind": "ConfigMap",
  "metadata": {
    "name": "test-configmap"
  },
  "data": {
    "test": "{{.testValue}}"
  }
}`)}

	psString = runtime.RawExtension{
		Raw: []byte(`{"type": "string"}`),
	}
)
