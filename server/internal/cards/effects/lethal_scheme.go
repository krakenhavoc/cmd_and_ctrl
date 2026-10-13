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
// "Each creature that convoked this spell" is the cast's own record
// (PaidCost.Convoked, #3172): the creatures tapped for convoke, in the
// order named. They connive one after another (ConniveEach), each
// completing its draw, discard and counter before the next begins; a
// convoker that has since left still draws and discards, with no
// counter. The destroy comes first, as printed. A copy of the spell was
// not cast, so nothing convoked it and nothing connives (CR 707.10).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4660b3c-a234-4a3e-83e7-e2fa9d556685",
		Name:         "Lethal Scheme",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyChosenPermanent(ctx.Game, item); err != nil {
				return err
			}
			return ConniveEach{Controller: item.Controller, Creatures: ctx.Paid().Convoked}.Apply(ctx)
		},
	})
}
