package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brazen Dwarf — Creature — Dwarf Shaman {1}{R}, 1/3 (Adventures in the
// Forgotten Realms):
//
//	"Whenever you roll one or more dice, this creature deals 1 damage to
//	 each opponent."
//
// One trigger per instruction to roll, however many dice (WheneverYouRollDice).
func init() {
	effects.Register(effects.Spec{
		OracleID: "256ab1e2-0c6a-41c1-ba01-021f3c90cfa2", Name: "Brazen Dwarf", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AfrWheneverYouRollDiceDo("Brazen Dwarf — 1 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}
