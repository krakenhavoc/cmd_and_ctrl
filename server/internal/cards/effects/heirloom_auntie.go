package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heirloom Auntie — Creature — Goblin Warlock {2}{B}, 4/4 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with two -1/-1 counters on it.
//	 Whenever another creature you control dies, surveil 1, then remove a
//	 -1/-1 counter from this creature."
//
// Brambleback Brute's entry replacement for the counters. The death
// trigger fires for each other creature of yours that dies, the Auntie's
// own death excluded and a second creature in the same batch a second
// trigger (the ruling, CR 603.10a). The counter comes off after the
// surveil is answered (Grave Researcher's tail), and only if the Auntie
// is still the same permanent and still has one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc9523da-f48e-49ab-9989-a4031dafb426",
		Name:         "Heirloom Auntie",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Heirloom Auntie: enters with two -1/-1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b39AnotherCreatureYouControlDied(ev, source, g)
			}, "Heirloom Auntie — surveil 1, then remove a -1/-1 counter from this creature", heirloomAuntieSurveil),
		},
	})
}

func heirloomAuntieSurveil(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return Surveil{N: 1, ThenTail: ctx.ItemTail(heirloomAuntieRemoveCounter)}.Apply(ctx)
}

// heirloomAuntieRemoveCounter is the "then remove a -1/-1 counter" half. An
// on-disk key.
var heirloomAuntieRemoveCounter = CardTail("heirloom-auntie/remove-counter", func(ctx *Context, _ game.TailOutcome) error {
	return stdB64RemoveAMinusCounterFromThis(ctx)
})
