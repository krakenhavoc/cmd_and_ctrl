package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_player_cards_test.go — #929, the three cards the
// resolution-time player choice shipped on.
//
// Slithermuse is the conversion (it auto-picked before), Gluntch is
// the chain of three exclusive choices with a sub-choice made by
// somebody who is not the controller, and Skullwinder is the
// cross-seat pair: the controller names an opponent and that opponent
// then picks a card.

const (
	GluntchOracle     = "0222dc7c-459b-4909-a037-72b2eb248599"
	SkullwinderOracle = "d3d8dd6e-a80b-4063-a542-ff67ef8dad4f"
)

// answerChoosePlayer answers `chooser`'s open player prompt by NAMING
// the seat, so a test says who it picked rather than guessing an
// index out of the prompt's bot ordering.
func AnswerChoosePlayer(t *testing.T, g *game.Game, chooser uuid.UUID, pick *game.Player) {
	t.Helper()
	c := LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no player prompt for %s: %+v", chooser, g.PendingChoices)
	}
	for i, opt := range c.PickOptions {
		if opt.Label != pick.Name {
			continue
		}
		if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
			t.Fatalf("ResolveOptionPick(%d): %v", i, err)
		}
		return
	}
	t.Fatalf("%q is not offered; options are %+v", pick.Name, c.PickOptions)
}

// --- helpers -------------------------------------------------------

// pushGraveyardCard puts a card straight into a player's graveyard.
func PushGraveyardCard(g *game.Game, p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// advanceToEndStep walks the cursor to the active player's end step,
// where Gluntch's trigger fires.
func AdvanceToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; g.Turn.Step != game.StepEnd; i++ {
		if i >= 12 {
			t.Fatalf("never reached the end step (stuck at %v)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}
