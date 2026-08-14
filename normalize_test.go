//nolint:testpackage // tests internal helpers
//go:build !goexperiment.jsonv2

package codec

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNormalizeForJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		check func(*testing.T, any)
	}{
		{
			name:  "nil",
			input: nil,
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				g.Expect(v).To(gomega.BeNil())
			},
		},
		{
			name:  "scalar string",
			input: "hello",
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				g.Expect(v).To(gomega.Equal("hello"))
			},
		},
		{
			name:  "scalar int",
			input: 42,
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				g.Expect(v).To(gomega.Equal(42))
			},
		},
		{
			name: "map[interface{}]interface{} with string keys",
			input: map[interface{}]interface{}{
				"name": "Alice",
				"age":  30,
			},
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				m, ok := v.(map[string]any)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(m[testFieldName]).To(gomega.Equal("Alice"))
				g.Expect(m["age"]).To(gomega.Equal(30))
			},
		},
		{
			name: "map[interface{}]interface{} with int keys",
			input: map[interface{}]interface{}{
				1: "one",
				2: "two",
			},
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				m, ok := v.(map[string]any)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(m["1"]).To(gomega.Equal("one"))
				g.Expect(m["2"]).To(gomega.Equal("two"))
			},
		},
		{
			name: "map[string]any passthrough",
			input: map[string]any{
				"key": "value",
			},
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				m, ok := v.(map[string]any)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(m[testMapKey]).To(gomega.Equal(testMapVal))
			},
		},
		{
			name:  "empty map[interface{}]interface{}",
			input: map[interface{}]interface{}{},
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				m, ok := v.(map[string]any)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(m).To(gomega.BeEmpty())
			},
		},
		{
			name:  "empty []any",
			input: []any{},
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				_, ok := v.([]any)
				g.Expect(ok).To(gomega.BeTrue())
			},
		},
		{
			name: "nested map[interface{}]interface{}",
			input: map[interface{}]interface{}{
				"outer": map[interface{}]interface{}{
					"inner": "value",
				},
			},
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				m, ok := v.(map[string]any)
				g.Expect(ok).To(gomega.BeTrue())
				inner, ok := m["outer"].(map[string]any)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(inner["inner"]).To(gomega.Equal(testMapVal))
			},
		},
		{
			name:  "[]any with mixed types",
			input: []any{1, "two", true},
			check: func(t *testing.T, v any) {
				g := gomega.NewWithT(t)
				s, ok := v.([]any)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(s).To(gomega.HaveLen(3))
				g.Expect(s[0]).To(gomega.Equal(1))
				g.Expect(s[1]).To(gomega.Equal("two"))
				g.Expect(s[2]).To(gomega.Equal(true))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeForJSON(tt.input)
			if err != nil {
				t.Fatalf("normalizeForJSON(%v) error: %v", tt.input, err)
			}
			tt.check(t, got)
		})
	}
}

func TestNormalizeForJSON_DepthCap(t *testing.T) {
	t.Parallel()

	// Build a deeply-nested map[interface{}]interface{} that exceeds maxNormalizeDepth.
	deep := make(map[interface{}]interface{})
	current := deep
	for range maxNormalizeDepth + 5 {
		next := make(map[interface{}]interface{})
		current["nested"] = next
		current = next
	}
	current["leaf"] = "bottom"

	_, err := normalizeForJSON(deep)
	if err == nil {
		t.Fatal("normalizeForJSON: expected depth-cap error, got nil")
	}
}

func TestNormalizeForJSON_AtMaxDepth(t *testing.T) {
	t.Parallel()

	// Build a map exactly at maxNormalizeDepth — should succeed.
	deep := make(map[interface{}]interface{})
	current := deep
	for range maxNormalizeDepth - 1 {
		next := make(map[interface{}]interface{})
		current["nested"] = next
		current = next
	}
	current["leaf"] = "bottom"

	_, err := normalizeForJSON(deep)
	if err != nil {
		t.Fatalf("normalizeForJSON: unexpected error at max depth: %v", err)
	}
}
