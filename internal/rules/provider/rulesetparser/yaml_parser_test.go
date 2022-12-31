package rulesetparser

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dadrus/heimdall/internal/rules/rule"
)

func TestParseYAML(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		uc     string
		conf   []byte
		assert func(t *testing.T, err error, ruleSet []rule.Configuration)
	}{
		{
			uc: "empty rule set spec",
			assert: func(t *testing.T, err error, ruleSet []rule.Configuration) {
				t.Helper()

				require.NoError(t, err)
				require.Empty(t, ruleSet)
			},
		},
		{
			uc:   "invalid rule set spec",
			conf: []byte(`- foo: bar`),
			assert: func(t *testing.T, err error, ruleSet []rule.Configuration) {
				t.Helper()

				require.Error(t, err)
			},
		},
		{
			uc:   "valid rule set spec",
			conf: []byte(`- id: bar`),
			assert: func(t *testing.T, err error, ruleSet []rule.Configuration) {
				t.Helper()

				require.NoError(t, err)
				require.Len(t, ruleSet, 1)
				assert.Equal(t, "bar", ruleSet[0].ID)
			},
		},
	} {
		t.Run(tc.uc, func(t *testing.T) {
			// WHEN
			ruleSet, err := parseYAML(bytes.NewBuffer(tc.conf))

			// THEN
			tc.assert(t, err, ruleSet)
		})
	}
}
