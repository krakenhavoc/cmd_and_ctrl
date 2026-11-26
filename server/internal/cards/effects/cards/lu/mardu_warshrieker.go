package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Warshrieker — Creature — Orc Shaman {3}{R}, 3/3 (Khans of Tarkir):
//
//	"Raid — When this creature enters, if you attacked this turn, add
//	 {R}{W}{B}."
//
// An intervening "if" (CR 603.4) on Priest of Urabrask's mana trigger: the
// mana is added as the trigger resolves and empties with the step (CR 500.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe9e9e2e-f304-4da3-a257-6886b0b1896b",
		Name:         "Mardu Warshrieker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB223RaidWhenThisEnters("Mardu Warshrieker — raid: add {R}{W}{B}",
				effects.Do(effects.AddMana{Produced: "{R}{W}{B}"})),
		},
	})
}
