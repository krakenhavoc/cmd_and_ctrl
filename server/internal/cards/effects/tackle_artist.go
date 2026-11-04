package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tackle Artist — Creature — Orc Sorcerer {3}{R}, 4/3 (Secrets of
// Strixhaven, #4456):
//
//	"Trample
//	 Opus — Whenever you cast an instant or sorcery spell, put a +1/+1
//	 counter on this creature. If five or more mana was spent to cast that
//	 spell, put two +1/+1 counters on this creature instead."
//
// Opus is an ability word (CR 207.2c); the ability is one trigger on your
// instant and sorcery casts (a copy is not cast, CR 707.10). The mana spent
// is counted as the spell is cast and carried on the item (Expressive
// Firedancer's Build), so the count holds even when the spell is countered
// before the ability resolves. "Instead" is one counter-placing event with
// either one or two counters, not one and then another.
//
// No simplification.
func init() {
	const label = "Tackle Artist — Opus: put a +1/+1 counter on this creature; with five or more mana spent, two instead"
	opus := On(game.EventCast, YouCast(Or(Instant(), Sorcery())), label,
		func(g *game.Game, item *game.StackItem) error {
			n := 1
			if item.Params.Amount >= 5 {
				n = 2
			}
			return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
		})
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "2e80d632-cf84-4562-8c93-5850cceb5bec",
		Name:            "Tackle Artist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{opus},
	})
}
