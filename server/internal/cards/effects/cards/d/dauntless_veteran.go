package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauntless Veteran — Creature — Human Soldier {1}{W}{W}, 2/2
// (Foundations, #4465):
//
//	"Whenever this creature attacks, creatures you control get +1/+1
//	 until end of turn."
//
// Pulse Tracker's attack trigger with Bearer of Glory's team pump. The
// creatures affected are the ones you control as the trigger resolves,
// the Veteran included, so one that enters later in the turn gets
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97b2fb8b-0c69-4ed2-b2d0-8a780f9b5b97",
		Name:         "Dauntless Veteran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Dauntless Veteran — creatures you control get +1/+1",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Dauntless Veteran — creatures you control get +1/+1",
				})),
		},
	})
}
