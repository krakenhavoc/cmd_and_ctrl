package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solstice Revelations — Instant — Lesson {2}{R}:
//
//	"Exile cards from the top of your library until you exile a nonland
//	 card. You may cast that card without paying its mana cost if the
//	 spell's mana value is less than the number of Mountains you
//	 control. If you don't cast that card this way, put it into your
//	 hand.
//	 Flashback {6}{R}"
//
// The walk is the exile-until-nonland run (Amped Raptor's), and the
// offer is Beseech the Mirror's (Game.CastExiledFreeOrToHandForEffect):
// a free cast made during the resolution (CR 608.2g), capped at one
// below the Mountains counted as the spell resolves, and the hand for a
// card that is declined or does not fit. With no Mountains the cap is
// below zero and the card goes straight to the hand. The lands exiled
// on the way stay in exile, as printed. An {X} in the card's cost is 0.
//
// Flashback is CR 702.34a: castable from the graveyard for {6}{R}, and
// exiled whenever it leaves the stack.
func init() {
	Register(Spec{
		OracleID:         "310c1280-dee8-426f-8440-6132a634a1e0",
		Name:             "Solstice Revelations",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t := ctx.ItemTail(solsticeRevelationsCastOrToHand)
			t.N = mountainsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Mountain")) - 1
			return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), 0, game.ZoneExile, UntilNonland, t)
		},
	})
}

// solsticeRevelationsCastOrToHand is the run's "then": the nonland card
// it exiled is offered free when its mana value is Tail.N or less, and
// put into its owner's hand otherwise. An on-disk key.
var solsticeRevelationsCastOrToHand = game.RegisterTail("solstice-revelations/cast-or-to-hand", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if err := g.CastExiledFreeOrToHandForEffect(o.Tail.Player, o.Tail.Source, id, o.Tail.N); err != nil {
			return err
		}
	}
	return nil
})
