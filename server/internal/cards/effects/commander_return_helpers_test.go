package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// commander_return_helpers_test.go — ADR 0115: a commander put into a
// graveyard or exile lands there like any other card, and the CR 903.9a
// state-based action then asks its OWNER whether it goes to the command
// zone (game.PendingChoiceCommanderReturn). These helpers find and answer
// that question. A bounce or a tuck is still CR 903.9b's replacement
// (game.PendingChoiceOptionalReplacement), answered before the card
// moves, and the b21/b29/b36 helpers answer that one.

// commanderReturnPromptFor returns the commander_return prompt owed to
// `owner`, or nil. The check runs at the next state-based action
// boundary, so a test that drove an effect directly (outside a
// resolution) may not have reached one yet: if nothing is open, the
// checks are run once and the prompts read again.
func commanderReturnPromptFor(g *game.Game, owner uuid.UUID) *game.PendingChoice {
	find := func() *game.PendingChoice {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceCommanderReturn && c.Chooser == owner {
				return c
			}
		}
		return nil
	}
	if c := find(); c != nil {
		return c
	}
	g.RunStateChecksForTest()
	return find()
}

// answerCommanderReturn answers CR 903.9a's question for `owner`: yes
// sends the commander from its graveyard or exile to the command zone,
// no leaves it where it is.
func answerCommanderReturn(t *testing.T, g *game.Game, owner uuid.UUID, apply bool) {
	t.Helper()
	c := commanderReturnPromptFor(g, owner)
	if c == nil {
		t.Fatalf("no commander_return prompt for %s", owner)
	}
	if err := g.ResolveCommanderReturn(c.ID, owner, apply); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
}
