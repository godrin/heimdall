// Copyright 2022 Dimitrij Drus <dadrus@gmx.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package extractors

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dadrus/heimdall/internal/heimdall"
	"github.com/dadrus/heimdall/internal/heimdall/mocks"
)

func TestExtractHeaderValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		uc             string
		strategy       HeaderValueExtractStrategy
		configureMocks func(t *testing.T, ctx *mocks.ContextMock)
		assert         func(t *testing.T, err error, authData AuthData)
	}{
		{
			uc:       "header is present, schema is irrelevant",
			strategy: HeaderValueExtractStrategy{Name: "X-Test-Header"},
			configureMocks: func(t *testing.T, ctx *mocks.ContextMock) {
				t.Helper()

				ctx.EXPECT().RequestHeader("X-Test-Header").Return("TestValue")
			},
			assert: func(t *testing.T, err error, authData AuthData) {
				t.Helper()

				assert.NoError(t, err)
				assert.Equal(t, "TestValue", authData.Value())
			},
		},
		{
			uc:       "schema is required, header is present, but without any schema",
			strategy: HeaderValueExtractStrategy{Name: "X-Test-Header", Schema: "Foo"},
			configureMocks: func(t *testing.T, ctx *mocks.ContextMock) {
				t.Helper()

				ctx.EXPECT().RequestHeader("X-Test-Header").Return("TestValue")
			},
			assert: func(t *testing.T, err error, authData AuthData) {
				t.Helper()

				assert.Error(t, err)
				assert.ErrorIs(t, err, heimdall.ErrArgument)
				assert.Contains(t, err.Error(), "'Foo' schema")
			},
		},
		{
			uc:       "schema is required, header is present, but with different schema",
			strategy: HeaderValueExtractStrategy{Name: "X-Test-Header", Schema: "Foo"},
			configureMocks: func(t *testing.T, ctx *mocks.ContextMock) {
				t.Helper()

				ctx.EXPECT().RequestHeader("X-Test-Header").Return("Bar TestValue")
			},
			assert: func(t *testing.T, err error, authData AuthData) {
				t.Helper()

				assert.Error(t, err)
				assert.ErrorIs(t, err, heimdall.ErrArgument)
				assert.Contains(t, err.Error(), "'Foo' schema")
			},
		},
		{
			uc:       "header with required schema is present",
			strategy: HeaderValueExtractStrategy{Name: "X-Test-Header", Schema: "Foo"},
			configureMocks: func(t *testing.T, ctx *mocks.ContextMock) {
				t.Helper()

				ctx.EXPECT().RequestHeader("X-Test-Header").Return("Foo TestValue")
			},
			assert: func(t *testing.T, err error, authData AuthData) {
				t.Helper()

				assert.NoError(t, err)
				assert.Equal(t, "TestValue", authData.Value())
			},
		},
		{
			uc:       "header is not present at all",
			strategy: HeaderValueExtractStrategy{Name: "X-Test-Header", Schema: "Foo"},
			configureMocks: func(t *testing.T, ctx *mocks.ContextMock) {
				t.Helper()

				ctx.EXPECT().RequestHeader("X-Test-Header").Return("")
			},
			assert: func(t *testing.T, err error, authData AuthData) {
				t.Helper()

				assert.Error(t, err)
				assert.ErrorIs(t, err, heimdall.ErrArgument)
				assert.Contains(t, err.Error(), "no 'X-Test-Header' header")
			},
		},
	} {
		t.Run("case="+tc.uc, func(t *testing.T) {
			// GIVEN
			ctx := mocks.NewContextMock(t)
			tc.configureMocks(t, ctx)

			// WHEN
			authData, err := tc.strategy.GetAuthData(ctx)

			// THEN
			tc.assert(t, err, authData)
		})
	}
}

func TestApplyHeaderAuthDataToRequest(t *testing.T) {
	t.Parallel()

	// GIVEN
	headerName := "X-Test-Header"
	rawHeaderValue := "Foo Bar"
	headerValueWithoutSchema := "Bar"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "foobar.local", nil)
	require.NoError(t, err)

	authData := &headerAuthData{name: headerName, rawValue: rawHeaderValue, value: headerValueWithoutSchema}

	// WHEN
	authData.ApplyTo(req)

	// THEN
	assert.Equal(t, rawHeaderValue, req.Header.Get(headerName))
}