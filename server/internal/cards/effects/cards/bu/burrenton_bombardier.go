package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burrenton Bombardier — Creature — Kithkin Soldier {2}{W}, 2/2
// (Morningtide, #5592):
//
//	"Flying
//	 Reinforce 2—{2}{W} ({2}{W}, Discard this card: Put two +1/+1 counters
//	 on target creature.)"
//
// Printed flying, and the shared reinforce row (morReinforce), which works
// only while the card is in its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13ec2390-e08a-4645-99e9-bac7e41917bb",
		Name:            "Burrenton Bombardier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.MorReinforce(2, "{2}{W}", "Reinforce 2—{2}{W} ({2}{W}, Discard this card: Put two +1/+1 counters on target creature.)"),
		},
	})
}
