package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aerial Doombot — {U} Artifact Creature — Robot Villain 1/1 (Marvel
// Super Heroes, #4455):
//
//	"Flying
//	 Power-up — {5}{U}: Put three +1/+1 counters on this creature.
//	 (Activate each power-up ability only once. Reduce the cost by its
//	 mana cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193, game/power_up.go): the turn it
// entered, the ability costs {5}{U} less {U}, which is {5}.
//
// No simplification.
func init() {
	// Purpose set by hand: a fourth AnswerPump caller of powerUpAnswering
	// trips unparam (#4571 drops that parameter).
	ab := effects.PowerUp("{5}{U}: Put three +1/+1 counters on this creature.",
		effects.ManaCost("{5}{U}"), effects.PlusOneCountersOnThis(3))
	ab.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:        "f24efc60-fdac-4a5d-bcbc-784ec5ae6d96",
		Name:            "Aerial Doombot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{ab},
	})
}
