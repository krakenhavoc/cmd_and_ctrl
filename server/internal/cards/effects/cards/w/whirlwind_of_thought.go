package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whirlwind of Thought — Enchantment {1}{U}{R}{W} (EDHREC rank 1627):
//
//	"Whenever you cast a noncreature spell, draw a card."
//
// The Jeskai spellslinger's engine: every noncreature spell
// replaces itself. Firebrand Archer's condition
// (b10NoncreatureSpellCastByYou — the spell is read off the stack,
// where its type line is intact) with a draw for the payoff. Fires
// once per cast, and the trigger goes on the stack above the spell
// so the card is drawn before the spell resolves, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6467cbb7-1e4e-482d-a20f-6cb9fc0f1ad1",
		Name:         "Whirlwind of Thought",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Whirlwind of Thought — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
