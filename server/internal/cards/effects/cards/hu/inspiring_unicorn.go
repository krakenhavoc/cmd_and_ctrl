package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Unicorn — Creature — Unicorn {2}{W}{W}, 2/2 (Guilds of Ravnica, #5525):
//
//	"Whenever this creature attacks, creatures you control get +1/+1 until
//	 end of turn."
//
// Dauntless Veteran's attack trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f9797e2-7ebb-41f9-9199-e9ee575eefd3",
		Name:         "Inspiring Unicorn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Inspiring Unicorn — creatures you control get +1/+1",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Inspiring Unicorn — creatures you control get +1/+1",
				})),
		},
	})
}
