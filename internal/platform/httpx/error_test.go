package httpx_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func TestError(t *testing.T) {
	err := errors.New("test")

	e := httpx.NewError(http.StatusServiceUnavailable, httpx.Code("test_code"), "Test msg", err)

	require.ErrorIs(t, e, err)
	assert.Equal(t, "test_code: Test msg: test", e.Error())
}

func TestMatchFieldErrors(t *testing.T) {
	err1 := errors.New("filed error 1")
	err2 := errors.New("filed error 2")
	err3 := errors.New("filed error 3")
	err := errors.Join(err1, err2, err3)
	rules := []httpx.FieldRule{
		{Field: "f1", Code: httpx.FieldCode("first"), Sentinel: err1},
		{Field: "f1", Code: httpx.FieldCode("second"), Sentinel: err2},
		{Field: "f2", Code: httpx.FieldCode("first"), Sentinel: err3},
	}
	want := []httpx.FieldError{
		{Field: "f1", Code: httpx.FieldCode("first")},
		{Field: "f2", Code: httpx.FieldCode("first")},
	}

	fe := httpx.MatchFieldErrors(err, rules)

	require.NotEmpty(t, fe)
	require.ElementsMatch(t, want, fe)
}
