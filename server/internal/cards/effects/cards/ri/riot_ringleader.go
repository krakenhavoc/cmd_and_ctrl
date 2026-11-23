package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riot Ringleader — {2}{R} Creature — Human Warrior 2/2 (Avacyn Restored, #5570):
//
//	"Whenever this creature attacks, Human creatures you control get
//	 +1/+0 until end of turn."
//
// Wandering Musicians' attack trigger, narrowed to Humans. The set
// is read as the ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "420d09b5-aced-4f71-914c-90d19f7476dc",
		Name:         "Riot Ringleader",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Riot Ringleader — Human creatures you control get +1/+0 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Human")), Power: 1, Label: "Riot Ringleader — +1/+0"})),
		},
	})
}
