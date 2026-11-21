package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multi_target_test.go — S20 sub-PR 5: clauses with more than one
// target slot.

const (
	ArcTrailOracle          = "f1c26b25-371e-4fbf-a43d-7fd59a364d3a"
	AshesToAshesOracle      = "944a52d9-bb14-43c5-8d05-2afaf023dc9f"
	SylvanReclamationOracle = "aeec8e85-6571-4da6-8a48-f5d3985ca10b"
)

func CardRefs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}
