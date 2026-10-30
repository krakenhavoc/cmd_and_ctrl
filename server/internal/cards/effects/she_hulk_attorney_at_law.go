package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// She-Hulk, Attorney-at-Law — {2}{G/W} Legendary Creature — Gamma Advisor
// Hero 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance
//	 Power-up — {6}{G/W}: Put a +1/+1 counter on She-Hulk. Then double the
//	 number of +1/+1 counters on each creature you control. (Activate each
//	 power-up ability only once. Reduce the cost by her mana cost if she
//	 entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The doubling is Bristly Bill's
// (b08DoubleCountersOnEachCreatureYouControl); the counter lands first, so
// She-Hulk's own new counter is doubled too.
//
// No simplification.
func init() {
	ab := PowerUp("{6}{G/W}: Put a +1/+1 counter on She-Hulk. Then double the number of +1/+1 counters on each creature you control.",
		ManaCost("{6}{G/W}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return b08DoubleCountersOnEachCreatureYouControl(g, item)
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "db99d52b-bbe6-4a48-8f2c-bf77912dc0a9",
		Name:            "She-Hulk, Attorney-at-Law",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []ActivatedAbility{ab},
	})
}
