package stash

import (
	"context"

	"github.com/nxenv/rapidship/nxenv-migrate/internal/codeerror"
)

func (e *Export) PullRequestReviewers(
	context.Context,
	int) error {
	return &codeerror.OpNotSupportedError{Name: "pullreqreview"}
}
