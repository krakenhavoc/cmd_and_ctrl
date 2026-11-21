package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathodion — Artifact Creature — Construct {3}, 3/3 (Urza's Saga):
//
//	"When this creature dies, add {C}{C}{C}."
//
// A dies trigger, not a mana ability: it does not trigger from a mana
// ability (CR 605.1b), so it uses the stack and the mana goes to the
// Construct's last controller as it resolves. The mana empties from the
// pool at the end of the step like any other (CR 106.4), so a creature
// that dies in combat gives mana for that combat step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcd4f816-2de1-4b30-82fb-cb87f45747ea",
		Name:         "Cathodion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Cathodion — add {C}{C}{C}", effects.Do(effects.AddMana{Produced: "{C}{C}{C}"})),
		},
	})
}
