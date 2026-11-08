package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bonecache Overseer — Creature — Squirrel Warlock {B}, 1/1
// (Bloomburrow, #4467):
//
//	"{T}, Pay 1 life: Draw a card. Activate only if three or more cards
//	 left your graveyard this turn or if you've sacrificed a Food this
//	 turn."
//
// The activation condition (CR 602.5b) is either half:
//
//   - three or more cards left your graveyard this turn, counted off the
//     turn's events (cardThatLeftAGraveyard: a zone move out of a
//     graveyard or a cast from one, tokens excluded). Each departure
//     counts, so a card that leaves, comes back and leaves again counts
//     twice, as the ruling says;
//   - you sacrificed a Food this turn, read off the per-turn sacrifice
//     tally (b34YouSacrificedAFoodThisTurn, Elanor Gardner's read), which
//     counts any Food artifact, not only a token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be3c7e4c-81fd-49e7-8fb5-5eb83994dafd",
		Name:         "Bonecache Overseer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}, Pay 1 life: Draw a card. Activate only if three or more cards left your graveyard this turn or if you've sacrificed a Food this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:      Plus(TapCost(), PayLife(1)),
			Condition: bonecacheOverseerCanActivate,
			Effect:    Do(DrawCards{N: 1}),
		}},
	})
}

// bonecacheOverseerCanActivate is the "Activate only if" clause.
func bonecacheOverseerCanActivate(g *game.Game, controller, _ uuid.UUID) bool {
	if b34YouSacrificedAFoodThisTurn(g, controller) {
		return true
	}
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if c, ok := cardThatLeftAGraveyard(ev, g); ok && c.Owner == controller {
			n++
		}
	}
	return n >= 3
}
