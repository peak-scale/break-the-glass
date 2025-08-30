package items

import (
	"bytes"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

func ValidateItems(tis TemplateItems) error {
	for name, ti := range tis {
		if _, err := ValidateSchema(ti.ParamSchema); err != nil {
			return fmt.Errorf("paramSchema for item %q is invalid: %w", name, err)
		}
		if _, err := ValidateTemplate(ti.Item); err != nil {
			return fmt.Errorf("template for item %q is invalid: %w", name, err)
		}
	}
	return nil
}

func Validate(ps ParamSchema, params Params) error {
	schema, err := ValidateSchema(ps)
	if err != nil || schema == nil {
		return err
	}

	// Create validator
	validator := validate.NewSchemaValidator(schema, nil, "", strfmt.Default)

	obj := params.Object
	if obj == nil {
		obj = make(map[string]any)
	}

	// Validate the data
	result := validator.Validate(obj)
	if !result.IsValid() {
		var errors []string
		for _, err := range result.Errors {
			errors = append(errors, err.Error())
		}
		return fmt.Errorf("validation failed: %v", errors)
	}
	return nil
}

// ValidateSchema prepares the validation schema. Returns nil if the schema is empty.
func ValidateSchema(ps ParamSchema) (*spec.Schema, error) {
	if ps.Empty() {
		return nil, nil
	}

	schemaData := ps.JSON()

	err := metaValidateJSONSchema(schemaData)
	if err != nil {
		return nil, fmt.Errorf("failed to validate OpenAPI schema: %w", err)
	}

	// Convert to OpenAPI spec schema
	schema := &spec.Schema{}
	if err := schema.UnmarshalJSON(schemaData); err != nil {
		return nil, fmt.Errorf("failed to create OpenAPI schema: %w", err)
	}
	return schema, nil
}

func metaValidateJSONSchema(schemaBytes []byte) error {
	// For OAS 3.1: https://json-schema.org/draft/2020-12/schema
	meta := "https://json-schema.org/draft/2020-12/schema"

	c := jsonschema.NewCompiler()
	if err := c.AddResource("meta.json", bytes.NewReader([]byte(`{"$ref":"`+meta+`"}`))); err != nil {
		return err
	}
	// Compile the candidate schema using the chosen meta-schema
	if err := c.AddResource("candidate.json", bytes.NewReader(schemaBytes)); err != nil {
		return err
	}
	if _, err := c.Compile("candidate.json"); err != nil {
		return fmt.Errorf("schema invalid: %w", err)
	}
	return nil
}
