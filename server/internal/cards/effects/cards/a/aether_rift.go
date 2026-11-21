package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Rift — Enchantment {1}{R}{G}:
//
//	"At the beginning of your upkeep, discard a card at random. If you
//	 discard a creature card this way, return it from your graveyard to
//	 the battlefield unless any player pays 5 life."
//
// ADR 0150 §3 (#1903): the 5 life is offered to every player in APNAP
// order from the active player, the Rift's controller first (it is their
// upkeep), until one pays. A player with less than 5 life, or whose life
// total can't change, can't pay (CR 119.4, 119.8). The offer holds the
// upkeep until it is answered.
//
// The offer is made only when the discarded card is a creature card
// that reached the graveyard. One discarded into exile by madness, or
// sent elsewhere by a replacement, can't be returned "from your
// graveyard", so there is nothing to pay to stop; the same holds if it
// leaves the graveyard while the offer is open.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2577ef8e-d85e-47af-ac34-629f31992751",
		Name:         "Aether Rift",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Aether Rift — discard a card at random", aetherRiftUpkeep),
		},
	})
}

func aetherRiftUpkeep(g *game.Game, item *game.StackItem) error {
	return g.DiscardRandomKeyedForEffect(item.Controller, 1, effects.NewContext(g, item).ItemTail(aetherRiftDiscarded))
}

// aetherRiftCreatureInGraveyard reports whether the card is a creature
// card in a graveyard.
func aetherRiftCreatureInGraveyard(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.IsCreature()
}

// aetherRiftReturns is "if no one does", keyed (#4264): the card that was discarded rides the offer.
var AetherRiftReturns = effects.AnyPlayerConsequences("any-player/aether-rift-return",
	nil,
	func(ctx *effects.Context, r game.AnyPlayerAnswer) error {
		if len(r.Carry) == 0 || !aetherRiftCreatureInGraveyard(ctx.Game, r.Carry[0]) {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: r.Carry[0], Dest: game.ZoneBattlefield}.Apply(ctx)
	},
	nil)

// aetherRiftDiscarded is the pay-or-return offer after the random discard (#4253 slice 9).
var aetherRiftDiscarded = effects.CardTail("aether-rift/offer-return-of-discarded-creature", func(ctx *effects.Context, o game.TailOutcome) error {
	g, item, discarded := ctx.Game, ctx.Item, o.IDs

	if len(discarded) == 0 || !aetherRiftCreatureInGraveyard(g, discarded[0]) {
		return nil
	}
	card := discarded[0]
	return effects.AnyPlayerPays{
		Pay:      effects.LifePayment(5),
		Question: "Aether Rift — pay 5 life to stop the creature card returning?",
		Then:     AetherRiftReturns,
		Carry:    []uuid.UUID{card},
	}.Apply(effects.NewContext(g, item))
})
