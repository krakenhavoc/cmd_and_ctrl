package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altered Ego — Creature — Shapeshifter {X}{2}{G}{U}, 0/0:
//
//	"This spell can't be countered.
//	 You may have this creature enter as a copy of any creature on the
//	 battlefield, except it enters with X additional +1/+1 counters on
//	 it."
//
// Clone's entry copy (EntersAsCopyOf over any creature) with Spark
// Double's kind of exception: the counters are not a copiable value,
// they change how the permanent enters, so they ride the entry event
// (AddCounterAtETB) and a counter replacement sees them. X is the one
// announced for the spell that became this permanent
// (EntryCastCountsForEffect, CR 107.3m); a permanent that was not cast
// has X of zero. Declining the copy leaves a 0/0 with no counters, which
// dies, as printed. The spell's own "can't be countered" is the printed
// Spec flag.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c35f3fd-c64e-4944-a4d5-37ce916d23c3",
		Name:            "Altered Ego",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Altered Ego", anyCreatureOnBattlefield,
				func(ev *game.ReplacementEvent, _ *game.PrintedValues, g *game.Game, _ *game.Card) {
					if x := g.EntryCastCountsForEffect(ev).X; x > 0 {
						ev.AddCounterAtETB(game.CounterPlusOne, x)
					}
				}),
		},
	})
}
