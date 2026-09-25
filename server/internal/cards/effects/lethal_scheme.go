package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lethal Scheme — Instant {2}{B}{B}:
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Destroy target creature or planeswalker. Each creature that
//	 convoked this spell connives. (Draw a card, then discard a card.
//	 If you discarded a nonland card, put a +1/+1 counter on that
//	 creature.)"
//
// Convoke is Spec.TapCost (S22); the destroy clause is the ordinary
// creature-or-planeswalker statement Hero's Downfall and Dreadbore
// both use.
//
// Declared simplification, weaker than printed: CONNIVE (CR 702.146)
// is not implemented anywhere in the engine — no primitive draws,
// discards and conditionally counters a specific creature as one
// unit — so the convoking creatures get none of it. The destroy half
// is unaffected and is the whole reason the card is cast.
func init() {
	Register(Spec{
		OracleID:     "c4660b3c-a234-4a3e-83e7-e2fa9d556685",
		Name:         "Lethal Scheme",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Creatures that convoked this spell don't connive — draw a card, discard a card, and get a +1/+1 counter for discarding a nonland card — that ability isn't implemented.",
		},
		TapCost: Convoke(),
		Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
}
