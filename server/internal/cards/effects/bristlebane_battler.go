package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bristlebane Battler — Creature — Kithkin Soldier {1}{G}, 6/6 (Lorwyn
// Eclipsed, #4458):
//
//	"Trample, ward {2}
//	 This creature enters with five -1/-1 counters on it.
//	 Whenever another creature you control enters while this creature
//	 has a -1/-1 counter on it, remove a -1/-1 counter from this
//	 creature."
//
// A 1/1 that grows by one each time another creature enters under your
// control, up to 6/6. "While this creature has a -1/-1 counter on it" is
// part of the trigger event, checked as the other creature enters, not an
// intervening if: with no -1/-1 counter left it doesn't trigger at all.
// The trigger then removes one counter if it still has one, so two
// creatures entering at once remove two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "592689f2-f17c-40d6-b880-431f4b6a6b0c",
		Name:            "Bristlebane Battler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 5, "Bristlebane Battler: enters with five -1/-1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Bristlebane Battler — ward {2}"),
			On(game.EventETB, AllOf(AnotherCreatureEnteredUnderYourControl, thisHasAMinusOneCounter),
				"Bristlebane Battler — remove a -1/-1 counter from this creature",
				removeAMinusOneCounterFromThis),
		},
	})
}

// thisHasAMinusOneCounter is "while this creature has a -1/-1 counter on
// it", read off the source permanent as the event happens.
func thisHasAMinusOneCounter(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.Counters[game.CounterMinusOne] > 0
}

// removeAMinusOneCounterFromThis removes one -1/-1 counter from the
// source, if it is still the permanent the ability came from (AddCounter
// asks) and still has one.
func removeAMinusOneCounterFromThis(g *game.Game, item *game.StackItem) error {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Counters[game.CounterMinusOne] <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterMinusOne, N: -1}.Apply(NewContext(g, item))
}
