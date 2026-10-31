package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kang Prime — {3}{U}{B} Legendary Creature — Human Villain 3/5 (Marvel
// Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Kang Prime enters or attacks, exile cards from the top of your
//	 library until you exile a nonland card. Put two time counters on that
//	 card. If it doesn't have suspend, it gains suspend. (At the beginning
//	 of your upkeep, remove a time counter. When the last is removed, you
//	 may cast it without paying its mana cost. If it's a creature, it has
//	 haste.)"
//
// One triggered ability with two events (OnAny). The walk is the shared
// exile-until-nonland run (MillToZone, a replacement window per card, a
// commander offered the command zone); its continuation puts the two time
// counters on the nonland card and gives suspend to one that has none
// (game.GainSuspendForEffect, CR 702.62a). The lands exiled on the way stay
// in exile. A library with no nonland card exiles all of itself and counts
// nothing down.
//
// What "gains suspend" does is the engine's: the card becomes one whose two
// exile-zone suspend triggers exist (the upkeep countdown and the free cast
// with haste on a creature), exactly as for a printed suspend card, and it
// stops having them the moment it leaves exile (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8e0b90c-edfb-4024-9ed5-398966f4ac72",
		Name:            "Kang Prime",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Kang Prime — exile until a nonland card, put two time counters on it, it gains suspend",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    UntilNonland,
						ThenTail: ctx.Tail(kangPrimeCounted),
					}.Apply(ctx)
				}),
		},
	})
}

// kangPrimeCounted is the clause after the exile: the nonland card that
// ended the run gets two time counters and, if it has none, suspend.
var kangPrimeCounted = CardTail("kang-prime/time-counters", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if err := ctx.Game.GainSuspendForEffect(ctx.Controller(), id, 2); err != nil {
			return err
		}
	}
	return nil
})
