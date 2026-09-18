package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/link/linktest"
	"github.com/iamroockie/linkforge/internal/link/usecase"
)

func TestCreateLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	saver := NewMockLinkSaver(ctrl)
	saver.EXPECT().Save(t.Context(), gomock.Any())
	uc := usecase.NewCreateLink(saver)

	got, err := uc.Create(t.Context(), linktest.CreateLinkParams())

	require.NoError(t, err)
	assert.NotZero(t, got)
	assert.NotEmpty(t, got.Alias)
}

func TestCreateLink_InvalidCreateLinkParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	saver := NewMockLinkSaver(ctrl)
	saver.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
	uc := usecase.NewCreateLink(saver)

	got, err := uc.Create(t.Context(), link.CreateLinkParams{})

	require.Error(t, err)
	require.ErrorContains(t, err, "create link")
	assert.Zero(t, got)
}

func TestCreateLink_Retries_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	saver := NewMockLinkSaver(ctrl)
	saver.EXPECT().Save(t.Context(), gomock.Any()).Return(link.ErrAliasTaken).Times(4)
	saver.EXPECT().Save(t.Context(), gomock.Any()).Return(nil)
	uc := usecase.NewCreateLink(saver)

	got, err := uc.Create(t.Context(), linktest.CreateLinkParams())

	require.NoError(t, err)
	assert.NotZero(t, got)
}

func TestCreateLink_Retries_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	saver := NewMockLinkSaver(ctrl)
	saver.EXPECT().Save(t.Context(), gomock.Any()).Return(link.ErrAliasTaken).Times(5)
	uc := usecase.NewCreateLink(saver)

	got, err := uc.Create(t.Context(), linktest.CreateLinkParams())

	require.ErrorContains(t, err, "alias taken")
	assert.Zero(t, got)
}

func TestCreateLink_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	saver := NewMockLinkSaver(ctrl)
	repoErr := errors.New("repo error")
	saver.EXPECT().Save(t.Context(), gomock.Any()).Return(repoErr)
	uc := usecase.NewCreateLink(saver)

	got, err := uc.Create(t.Context(), linktest.CreateLinkParams())

	require.ErrorIs(t, err, repoErr)
	assert.Zero(t, got)
}
