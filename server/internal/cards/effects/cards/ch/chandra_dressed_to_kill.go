package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Dressed to Kill — Legendary Planeswalker — Chandra {1}{R}{R},
// loyalty 3 (Innistrad: Crimson Vow):
//
//	"+1: Add {R}. Chandra deals 1 damage to up to one target player or
//	     planeswalker.
//	 +1: Exile the top card of your library. If it's red, you may cast it
//	     this turn.
//	 −7: Exile the top five cards of your library. You may cast red spells
//	     from among them this turn. You get an emblem with "Whenever you
//	     cast a red spell, this emblem deals X damage to any target, where
//	     X is the amount of mana spent to cast that spell.""
//
// All three are loyalty abilities, so they use the stack and are not mana
// abilities (CR 605.1a): the first adds {R} as it resolves (the mana empties
// with the step), and the damage's "up to one" target may be left empty.
//
// The second +1 exiles the top card through the keyed exile, and only when
// it is red does it open a this-turn permission to CAST it (a paid cast, at
// the card's own timing; a land is never red). A card that is not red stays
// exiled with no permission.
//
// The −7 exiles the top five, and opens one permission over the red cards
// among them (CR 611.2c locks the set; any number may be cast, each paid
// for), then gives the emblem. The emblem triggers on each red spell its
// controller casts and deals damage equal to the mana spent on that spell,
// read as the spell is cast and carried on the item, so a spell that is
// countered in response still pays (a spell cast for no mana deals 0). The
// emblem is the source of its own damage.
//
// No simplification.
func init() {
	const emblemLabel = "Chandra, Dressed to Kill emblem — X damage to any target, X the mana spent to cast that spell"
	emblem := effects.Targeting(
		effects.WheneverYouCast(effects.OfColor("R"), emblemLabel, chandraDressedEmblemDamage),
		effects.TargetAny())
	emblem.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, emblemLabel)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "4686f776-2055-470c-bd99-7c4bb32902c0",
		Name:            "Chandra, Dressed to Kill",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		ReadsManaSpent:  game.ReadsYourCasts,
		Emblem: &effects.EmblemSpec{
			Label:     "Chandra, Dressed to Kill emblem",
			Text:      "Whenever you cast a red spell, this emblem deals X damage to any target, where X is the amount of mana spent to cast that spell.",
			Triggered: []game.TriggeredAbility{emblem},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Add {R}. Chandra deals 1 damage to up to one target player or planeswalker.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPlayerOrPlaneswalker().WithCount(0, 1),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.AddMana{Produced: "{R}"}).Apply(ctx); err != nil {
						return err
					}
					return effects.SourceDealsDamageToEachLegalTarget(1)(g, item)
				},
			},
			{
				Label: "+1: Exile the top card of your library. If it's red, you may cast it this turn.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 1, ctx.ItemTail(chandraDressedCastIfRed))
				},
			},
			{
				Label: "−7: Exile the top five cards of your library. You may cast red spells from among them this turn. You get an emblem with \"Whenever you cast a red spell, this emblem deals X damage to any target, where X is the amount of mana spent to cast that spell.\"",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 5, ctx.ItemTail(chandraDressedUltimate))
				},
			},
		},
	})
}

// chandraDressedGrant opens the this-turn permission to cast the red cards
// among `ids` that are in exile, one permission over the set. Not red, not
// in exile: not granted. One permission serves the −7 (any number may be
// cast) and the +1 (one card).
func chandraDressedGrant(g *game.Game, controller, source uuid.UUID, label string, ids []uuid.UUID) {
	var cards []game.Card
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.HasColorWord("R") {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		cards = append(cards, c)
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     controller,
		Zone:       game.ZoneExile,
		Duration:   g.UntilEndOfTurnDuration(),
		CastOnly:   true,
		Source:     source,
		SourceName: "Chandra, Dressed to Kill",
		Label:      label,
	}, cards)
}

// chandraDressedCastIfRed is the first exile's rest: "If it's red, you may
// cast it this turn". An on-disk key: never renamed, never reused.
var chandraDressedCastIfRed = effects.CardTail("chandra-dressed-to-kill/cast-if-red", func(ctx *effects.Context, o game.TailOutcome) error {
	chandraDressedGrant(ctx.Game, ctx.Controller(), ctx.Source(), "Chandra, Dressed to Kill — cast the exiled red card this turn", o.IDs)
	return nil
})

// chandraDressedUltimate is the −7's rest: the permission over the red
// cards among those exiled, then the emblem. An on-disk key.
var chandraDressedUltimate = effects.CardTail("chandra-dressed-to-kill/ultimate", func(ctx *effects.Context, o game.TailOutcome) error {
	chandraDressedGrant(ctx.Game, ctx.Controller(), ctx.Source(), "Chandra, Dressed to Kill — cast a red spell from among the exiled cards this turn", o.IDs)
	return effects.CreateEmblem{}.Apply(ctx)
})

// chandraDressedEmblemDamage is the emblem trigger's resolution: X damage,
// X the mana spent on the spell as it was cast, from the emblem to the
// target chosen as it triggered.
func chandraDressedEmblemDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if item.Params.Amount <= 0 {
		return nil
	}
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: item.Params.Amount}.Apply(ctx)
}
