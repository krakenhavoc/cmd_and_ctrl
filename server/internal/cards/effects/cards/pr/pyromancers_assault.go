package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyromancer's Assault — Enchantment {3}{R} (Oath of the Gatewatch, #5545):
//
//	"Whenever you cast your second spell each turn, this enchantment deals 2
//	 damage to any target."
//
// The trigger is YouCastYourSecondSpellEachTurn (Jori En's count): the second
// spell, whatever it is, however many were cast before the Assault arrived. The
// target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87038d1b-0fa9-4604-b119-1de56db30144",
		Name:         "Pyromancer's Assault",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
					"Pyromancer's Assault — 2 damage to any target", effects.B33DamageChosenTargetFromSource(2)),
				effects.TargetAny()),
		},
	})
}
