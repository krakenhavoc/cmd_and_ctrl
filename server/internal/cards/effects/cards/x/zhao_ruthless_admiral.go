package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zhao, Ruthless Admiral — Legendary Creature — Human Soldier
// {2}{B/R}{B/R}, 3/4 (Avatar: The Last Airbender, #4459):
//
//	"Firebending 2 (Whenever this creature attacks, add {R}{R}. This mana
//	 lasts until end of combat.)
//	 Whenever you sacrifice another permanent, creatures you control get
//	 +1/+0 until end of turn."
//
// Firebending is Avatar Roku's (`firebending`). The sacrifice trigger is
// Pirate Peddlers' condition (avB18YouSacrificedAnother), one trigger per
// permanent sacrificed, and the pump is Master of Barbs'.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6b0198a-c1f2-4142-b5e2-8205adf14d0a",
		Name:         "Zhao, Ruthless Admiral",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Zhao, Ruthless Admiral — firebending 2", effects.Do(effects.NewFirebending(2))),
			effects.On(game.EventSacrifice, effects.AvB18YouSacrificedAnother,
				"Zhao, Ruthless Admiral — creatures you control get +1/+0 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Zhao, Ruthless Admiral — +1/+0"})),
		},
	})
}
