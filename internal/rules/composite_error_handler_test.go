package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dadrus/heimdall/internal/pipeline/testsupport"
)

func TestCompositeErrorHandlerExecutionWithFallback(t *testing.T) {
	t.Parallel()

	// GIVEN
	ctx := &testsupport.MockContext{}
	ctx.On("AppContext").Return(context.Background())

	eh1 := &mockErrorHandler{}
	eh1.On("Execute", ctx, testsupport.ErrTestPurpose).Return(false, nil)

	eh2 := &mockErrorHandler{}
	eh2.On("Execute", ctx, testsupport.ErrTestPurpose).Return(true, nil)

	eh := compositeErrorHandler{eh1, eh2}

	// WHEN
	ok, err := eh.Execute(ctx, testsupport.ErrTestPurpose)

	// THEN
	assert.NoError(t, err)
	assert.True(t, ok)

	eh1.AssertExpectations(t)
	eh2.AssertExpectations(t)
}

func TestCompositeErrorHandlerExecutionWithoutFallback(t *testing.T) {
	t.Parallel()

	// GIVEN
	ctx := &testsupport.MockContext{}
	ctx.On("AppContext").Return(context.Background())

	eh1 := &mockErrorHandler{}
	eh1.On("Execute", ctx, testsupport.ErrTestPurpose).Return(true, nil)

	eh2 := &mockErrorHandler{}

	eh := compositeErrorHandler{eh1, eh2}

	// WHEN
	ok, err := eh.Execute(ctx, testsupport.ErrTestPurpose)

	// THEN
	assert.NoError(t, err)
	assert.True(t, ok)

	eh1.AssertExpectations(t)
	eh2.AssertExpectations(t)
}

func TestCompositeErrorHandlerExecutionWithNoApplicableErrorHandler(t *testing.T) {
	t.Parallel()

	// GIVEN
	ctx := &testsupport.MockContext{}
	ctx.On("AppContext").Return(context.Background())

	eh1 := &mockErrorHandler{}
	eh1.On("Execute", ctx, testsupport.ErrTestPurpose).Return(false, nil)

	eh2 := &mockErrorHandler{}
	eh2.On("Execute", ctx, testsupport.ErrTestPurpose).Return(false, nil)

	eh := compositeErrorHandler{eh1, eh2}

	// WHEN
	ok, err := eh.Execute(ctx, testsupport.ErrTestPurpose)

	// THEN
	assert.Error(t, err)
	assert.False(t, ok)

	eh1.AssertExpectations(t)
	eh2.AssertExpectations(t)
}