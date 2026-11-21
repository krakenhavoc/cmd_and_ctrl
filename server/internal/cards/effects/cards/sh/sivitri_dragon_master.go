package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sivitri, Dragon Master — Legendary Planeswalker — Sivitri {2}{U}{B},
// loyalty 4:
//
//	"+1: Until your next turn, creatures can't attack you or
//	 planeswalkers you control unless their controller pays 2 life for
//	 each of those creatures.
//	 −3: Search your library for a Dragon card, reveal it, put it into
//	 your hand, then shuffle.
//	 −7: Destroy all non-Dragon creatures.
//	 Sivitri, Dragon Master can be your commander."
//
// The +1 registers a duration attack tax paid in life (#5869,
// game/attack_tax_duration.go's ModAttackTaxLife): it names the
// controller, covers their planeswalkers as printed, and ends as their
// next turn begins. The attacking player pays 2 life per creature
// attacking the controller or one of their planeswalkers, only from a
// life total at least that large (CR 119.4).
//
// The −3 is Maelstrom of the Spirit Dragon's search. The −7 spares
// every Dragon, the controller's included, through the effective
// subtype. "Can be your commander" is printed card data the deck
// validator reads.
//
// No simplification.
func init() {
	const name = "Sivitri, Dragon Master"
	const taxClause = "Until your next turn, creatures can't attack you or planeswalkers you control unless their controller pays 2 life for each of those creatures."
	effects.Register(effects.Spec{
		OracleID:     "5cf86603-cc47-4644-890a-0997c694817f",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: " + taxClause,
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.AttackTaxLifeForEffect(item.SourceCardID, item.Controller, 2,
						game.AttackTaxOnPlayerOrPlaneswalkers,
						g.UntilYourNextTurnDuration(item.Controller), name+" — "+taxClause)
					return nil
				},
			},
			{
				Label:   "−3: Search your library for a Dragon card, reveal it, put it into your hand, then shuffle.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Tutors: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					dragon := ctx.TextSubtype("Dragon")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.HasSubtype(dragon) },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    name + " — a Dragon card",
					}.Apply(ctx)
				},
			},
			{
				Label: "−7: Destroy all non-Dragon creatures.",
				Cost:  effects.LoyaltyCost(-7),
				Purpose: game.Purpose{Sweep: game.Sweep{
					Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					dragon := ctx.TextSubtype("Dragon")
					return effects.DestroyAllMatching{
						Match: effects.And(effects.Creature(), effects.Not(effects.OfCreatureType(dragon))),
					}.Apply(ctx)
				},
			},
		},
	})
}
