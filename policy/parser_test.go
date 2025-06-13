package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultParser_Parse_ValidPolicy(t *testing.T) {
	parser := NewDefaultParser()
	policyJSON := []byte(`{
		"name": "test-policy",
		"description": "A test policy",
		"type": "load_balancing",
		"priority": 1,
		"conditions": {
			"all": [
				{"metric": "cpu_usage", "operator": "gt", "value": 80.0}
			]
		},
		"actions": [
			{"type": "migrate_shard"}
		]
	}`)

	policy, err := parser.Parse(policyJSON)
	assert.NoError(t, err)
	assert.NotNil(t, policy)
	assert.Equal(t, "test-policy", policy.Name)
	assert.Equal(t, PolicyTypeLoadBalancing, policy.Type)
	assert.Len(t, policy.Conditions.All, 1)
	assert.Len(t, policy.Actions, 1)
}

func TestDefaultParser_Parse_InvalidPolicy(t *testing.T) {
	parser := NewDefaultParser()
	policyJSON := []byte(`{
		"name": "",
		"type": "",
		"conditions": {},
		"actions": []
	}`)

	policy, err := parser.Parse(policyJSON)
	assert.Error(t, err)
	assert.Nil(t, policy)
}

func TestDefaultParser_Validate_MissingAction(t *testing.T) {
	parser := NewDefaultParser()
	p := &Policy{
		Name: "no-action",
		Type: PolicyTypeLoadBalancing,
		Conditions: Conditions{
			All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
		},
		Actions: []Action{},
	}
	err := parser.Validate(p)
	assert.Error(t, err)
}
