package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brave Brawler — {1}{W} Creature — Human Warrior Hero 2/1:
//
//	"Lifelink
//	 Power-up — {4}{W}: Put two +1/+1 counters on this creature.
//	 (Activate each power-up ability only once. Reduce the cost by its
//	 mana cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193, game/power_up.go): the turn it
// entered, the ability costs {4}{W} less {1}{W}, which is {3}; after
// that it costs {4}{W}; and it can be activated once per object.
//
// No simplification. Marvel Super Heroes (#4455, seam #4527).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a1a791be-71e6-4c1f-8717-08ca62d40de2",
		Name:            "Brave Brawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{
			effects.PowerUpAnswering("{4}{W}: Put two +1/+1 counters on this creature.",
				effects.ManaCost("{4}{W}"), effects.PlusOneCountersOnThis(2)),
		},
	})
}
