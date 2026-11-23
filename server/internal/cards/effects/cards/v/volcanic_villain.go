package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Villain — {2}{R} Creature — Elemental Villain 3/2 (Marvel
// Super Heroes, #4455):
//
//	"Haste
//	 Power-up — {5}{R}: Put two +1/+1 counters on this creature.
//	 (Activate each power-up ability only once. Reduce the cost by its
//	 mana cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193, game/power_up.go): the turn it
// entered, the ability costs {5}{R} less {2}{R}, which is {3}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a257426d-70c9-4d58-bd8b-97ec38f6e133",
		Name:            "Volcanic Villain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{
			effects.PowerUpAnswering("{5}{R}: Put two +1/+1 counters on this creature.",
				effects.ManaCost("{5}{R}"), effects.PlusOneCountersOnThis(2)),
		},
	})
}
