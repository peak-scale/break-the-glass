/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"errors"
	"fmt"

	"github.com/peak-scale/break-the-glass/internal/items"
	"k8s.io/apimachinery/pkg/runtime"
)

func (brt *BreakRequestTemplate) RenderItemsItems(br *BreakRequest) (items.Items, error) {
	params := br.Spec.Params
	if params == nil {
		params = items.TemplateParams{}
	}
	rendered := make(items.Items, len(brt.Spec.Items))

	var rerr error
	for name, i := range brt.Spec.Items {
		var p []byte
		if ip, ok := params[name]; ok {
			p = ip.Raw
		}
		r, err := items.RenderTemplate(i.ManifestTemplate.Raw, p)
		if err != nil {
			rerr = errors.Join(rerr, fmt.Errorf("error rendering template item %s: %w", name, err))
		}
		rendered[name] = &runtime.RawExtension{Raw: r}
	}
	return rendered, rerr
}
