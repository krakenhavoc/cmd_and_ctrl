package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timbermaw Larva — Creature — Beast {3}{G}, 2/2 (Zendikar, S82 OLD b496):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn for each Forest
//	 you control."
//
// The Forests are counted as the trigger resolves (Old496LarvaAttackPump); the
// bonus lasts to end of turn and goes only on the Larva that attacked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e53532db-5b14-429e-8437-a369c5b8632b",
		Name:         "Timbermaw Larva",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Timbermaw Larva — it gets +1/+1 until end of turn for each Forest you control", effects.Old496LarvaAttackPump),
		},
	})
}
