package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nobilis of War — Creature — Spirit Avatar {R/W}{R/W}{R/W}{R/W}{R/W}, 3/4 (Eventide, #5590):
//
//	"Flying
//	 Attacking creatures you control get +2/+0."
//
// Instigator Gang's layer 7c boost over your attacking creatures, the
// Nobilis itself included when it attacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4789ac35-ca91-4046-8262-24c34e751cea",
		Name:            "Nobilis of War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.AttackingCreaturesYouControlGetPower(2)},
	})
}
