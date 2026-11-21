package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fell Beast of Mordor — Creature — Drake Beast {2}{B}{B}, 3/3 (Tales of
// Middle-earth Commander, #5948):
//
//	"Flying
//	 Devour 1 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with that many +1/+1 counters on it.)
//	 Whenever this creature enters or attacks, target opponent loses X life
//	 and you gain X life, where X is the number of +1/+1 counters on it."
//
// Devour is the shared as-enters replacement (devour.go, CR 702.82a), so the
// counters are on the Beast before its enters trigger is put on the stack.
// The trigger is one printed ability on two events (WhenThisEntersOrAttacks)
// with a target opponent chosen as it goes on the stack; X is counted as it
// resolves, from the Beast on the battlefield or, if it has left, as it last
// existed there (CR 608.2h). A target that became illegal drains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b64e0a8-7d11-41c5-b439-edec19928b7d",
		Name:            "Fell Beast of Mordor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{Devour("Fell Beast of Mordor", 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("Fell Beast of Mordor — target opponent loses X life and you gain X life, where X is the number of +1/+1 counters on it",
					fellBeastDrain),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// fellBeastDrain is the trigger body: X is the +1/+1 counters on the source.
func fellBeastDrain(g *game.Game, item *game.StackItem) error {
	x := 0
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && onBattlefield(g, item.SourceCardID) {
		x = c.Counters[game.CounterPlusOne]
	} else {
		x = g.LastKnownCountersForEffect(item.SourceCardID)[game.CounterPlusOne]
	}
	if x <= 0 {
		return nil
	}
	return targetOpponentLosesAndYouGain(g, item, x)
}
