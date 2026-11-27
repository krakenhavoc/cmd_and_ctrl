package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Preyseizer Dragon — Creature — Dragon {4}{R}{R}, 4/4 (Planechase 2012, S82
// OLD, #5569):
//
//	"Flying
//	 Devour 2 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with twice that many +1/+1 counters on it.)
//	 Whenever this creature attacks, it deals damage to any target equal to the
//	 number of +1/+1 counters on this creature."
//
// Devour 2 is Predator Dragon's (CR 702.82a); the counters ride the entry event.
// The attack trigger counts +1/+1 counters as it resolves, or as the Dragon
// last existed if it has left (CR 608.2h), and nothing is dealt for none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b36eb85-0b1e-41fb-8ecd-dda13c4feffc",
		Name:            "Preyseizer Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.Devour("Preyseizer Dragon", 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Preyseizer Dragon — it deals damage to any target equal to the number of +1/+1 counters on it",
				effects.S82OldB432DamageEqualToCounters), effects.TargetAny()),
		},
	})
}
