package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vivisection Evangelist — Creature — Phyrexian Cleric {3}{W}{B}, 4/4:
//
//	"Vigilance
//	 Corrupted — When this creature enters, if an opponent has three or more
//	 poison counters, destroy target creature or planeswalker an opponent
//	 controls."
//
// An intervening "if" (CR 603.4): with no corruption as it enters, the
// trigger never goes on the stack (so no target is chosen); it is checked
// again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "007f3b20-d58b-4986-8de8-9e419af88628",
		Name:            "Vivisection Evangelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIf(effects.CondCorrupted{}, effects.On(game.EventETB, effects.AllOf(effects.Self, effects.YouAreCorrupted),
					"Corrupted — Vivisection Evangelist — destroy target creature or planeswalker an opponent controls",
					effects.DestroyFirstLegalTarget)),
				effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
		},
	})
}
