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
	"fmt"

	"github.com/google/cel-go/cel"
	"k8s.io/apimachinery/pkg/runtime"
)

func (brt *BreakRequestTemplate) IsApproved(br *BreakRequest) (bool, error) {
	if !brt.Spec.AutoApprove {
		return false, nil
	}
	if brt.Spec.ApprovalCondition == "" {
		return true, nil
	}

	env, err := cel.NewEnv(
		cel.Variable("request", cel.DynType),
	)
	if err != nil {
		return false, err
	}

	ast, iss := env.Compile(brt.Spec.ApprovalCondition)
	if iss != nil && iss.Err() != nil {
		return false, iss.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return false, err
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(br)
	if err != nil {
		return false, err
	}

	result, _, err := prg.Eval(map[string]interface{}{
		"request": obj,
	})

	// Convert the result to boolean
	boolResult, ok := result.Value().(bool)
	if !ok {
		return false, fmt.Errorf(
			"expression did not evaluate to a boolean, got: %T",
			result.Value(),
		)
	}
	return boolResult, err
}
