package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d6b0198a-c1f2-4142-b5e2-8205adf14d0a",
		Name:         "Zhao, Ruthless Admiral",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Zhao, Ruthless Admiral — firebending 2", Do(firebending(2))),
			On(game.EventSacrifice, avB18YouSacrificedAnother,
				"Zhao, Ruthless Admiral — creatures you control get +1/+0 until end of turn",
				Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Zhao, Ruthless Admiral — +1/+0"})),
		},
	})
}
