package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_after.go — "…. The Ring tempts you." after an action that
// can pause (ADR 0114 PR 5). Append-only.
//
// A destruction, a bounce, damage and a scry can each stop for a
// prompt (CR 903.9's command-zone question, a CR 616 replacement
// order, the scry's own reorder), and the sentence after them must
// not run until they are done: a commander still waiting for its
// owner's answer is on the battlefield, and must not be offered as the
// Ring-bearer. So the tempt goes in the action's continuation.

// ringTemptsYouNext is the tempt as a continuation: it runs against
// the game the continuation is handed (the StackItem.Effect contract,
// so an undo across the prompt lands on the restored game) and is
// controlled by the item's controller.
func ringTemptsYouNext(item *game.StackItem) func(*game.Game) error {
	return func(g *game.Game) error { return TheRingTemptsYou{}.Apply(NewContext(g, item)) }
}

// legalTargetIDs is the IDs of the item's targets that are still legal
// as it resolves (CR 608.2b), in announce order: the list a batch
// action takes.
func legalTargetIDs(ctx *Context) []uuid.UUID {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		ids = append(ids, t.ID)
	}
	return ids
}
