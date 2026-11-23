package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prowling Felidar — Creature — Cat Beast {3}{W}, 2/3 (Zendikar Rising,
// #5508):
//
//	"Vigilance
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter
//	 on this creature."
//
// Vigilance is printed; the landfall counter is Territorial Scythecat's
// (an ability word, CR 207.2c), a Do of CounterOnThis so a Felidar that
// left in response gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed26f6bf-fcd6-4cf4-9b65-a492cb79f2df",
		Name:            "Prowling Felidar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Prowling Felidar — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
