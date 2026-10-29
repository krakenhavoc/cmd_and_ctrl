package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f24efc60-fdac-4a5d-bcbc-784ec5ae6d96",
		Name:            "Aerial Doombot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			powerUpAnswering(game.AnswerPump, "{5}{U}: Put three +1/+1 counters on this creature.",
				ManaCost("{5}{U}"), plusOneCountersOnThis(3)),
		},
	})
}
