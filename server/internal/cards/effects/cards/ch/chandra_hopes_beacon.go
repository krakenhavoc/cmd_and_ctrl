package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, Hope's Beacon — Legendary Planeswalker — Chandra {4}{R}{R},
// starting loyalty 5 (March of the Machine):
//
//	"Whenever you cast an instant or sorcery spell, copy it. You may choose
//	 new targets for the copy. This ability triggers only once each turn.
//	 +2: Add two mana in any combination of colors.
//	 +1: Exile the top five cards of your library. Until the end of your
//	     next turn, you may cast an instant or sorcery spell from among
//	     those exiled cards.
//	 −X: Chandra deals X damage to each of up to two targets."
//
// The trigger is Thousand-Year Storm's copy of the spell just cast,
// once (CopySpell names the spell without targeting it, so a countered
// spell is still copied, CR 608.2h), under TriggersOnlyOnceEachTurn. The
// +2 is a loyalty ability, not a mana ability (CR 605.1a), so it goes on the
// stack; the two mana are each of a colour its controller names.
//
// The +1 exiles the five cards together, then opens ONE permission over the
// instants and sorceries among them (CastsLeft 1: "an instant or sorcery
// spell", CR 611.2c locks the set), lasting through the end of your next
// turn. The spell is cast at its own timing and paid for; a card that is not
// an instant or sorcery is stranded in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53231b31-fb3c-4b30-bfe9-0ee8e754247b",
		Name:            "Chandra, Hope's Beacon",
		Completeness:    effects.CompletenessFull,
		XMatters:        true,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{effects.TriggersOnlyOnceEachTurn(game.TriggeredAbility{
			Watches: []game.EventKind{game.EventCast},
			Key:     "Chandra, Hope's Beacon — copy the spell",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CopySpell{
					StackID:          item.Trigger.Event.CardID,
					Controller:       item.Controller,
					Count:            1,
					ChooseNewTargets: true,
					FromLastKnown:    true,
				}.Apply(effects.NewContext(g, item))
			},
		})},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Add two mana in any combination of colors.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AddMana{Produced: effects.AnyCombinationOfColors(2)}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label: "+1: Exile the top five cards of your library. Until the end of your next turn, you may cast an instant or sorcery spell from among those exiled cards.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 5, ctx.Tail(chandraHopesBeaconGrant))
				},
			},
			{
				Label:   "−X: Chandra deals X damage to each of up to two targets.",
				Cost:    effects.LoyaltyMinusX(),
				Targets: effects.TargetAny().WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if ctx.X() <= 0 {
						return nil
					}
					return effects.B10DamageEachLegalTarget(ctx, ctx.X())
				},
			},
		},
	})
}

// chandraHopesBeaconGrant is the +1's tail: one cast, of an instant or
// sorcery from among the exiled cards, until the end of your next turn. An
// on-disk key: never renamed, never reused.
var chandraHopesBeaconGrant = effects.CardTail("chandra-hopes-beacon/grant-cast", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	var cards []game.Card
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && (c.IsInstant() || c.IsSorcery()) {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   g.UntilEndOfYourNextTurnDuration(ctx.Controller()),
		CastOnly:   true,
		CastsLeft:  1,
		Filter:     game.PermissionFilter{InstantOrSorceryOnly: true},
		Source:     ctx.Source(),
		SourceName: "Chandra, Hope's Beacon",
		Label:      "Chandra, Hope's Beacon — cast an instant or sorcery spell from among the exiled cards",
	}, cards)
	return nil
})
