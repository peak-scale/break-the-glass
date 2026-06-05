package conditions

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/peak-scale/break-the-glass/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func IsApproved(brt *v1alpha1.BreakRequestTemplate, br *v1alpha1.BreakRequest) (bool, error) {
	if !brt.Spec.AutoApprove {
		return false, nil
	}
	if brt.Spec.ApprovalCondition == "" {
		return true, nil
	}

	prg, err := PrepareCondition(brt)
	if err != nil {
		return false, err
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(br)
	if err != nil {
		return false, err
	}

	result, _, err := prg.Eval(map[string]any{
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

func PrepareCondition(brt *v1alpha1.BreakRequestTemplate) (cel.Program, error) {
	env, err := cel.NewEnv(
		cel.Variable("request", cel.DynType),
	)
	if err != nil {
		return nil, err
	}

	ast, iss := env.Compile(brt.Spec.ApprovalCondition)
	if iss != nil && iss.Err() != nil {
		return nil, iss.Err()
	}
	return env.Program(ast)
}
