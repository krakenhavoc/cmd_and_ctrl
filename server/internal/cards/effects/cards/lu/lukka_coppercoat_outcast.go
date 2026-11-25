package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lukka, Coppercoat Outcast — Legendary Planeswalker — Lukka {3}{R}{R},
// loyalty 5 (Ikoria: Lair of Behemoths, S82 OLD batch 120, #5512):
//
//	"+1: Exile the top three cards of your library. Creature cards exiled this
//	 way gain "You may cast this card from exile as long as you control a Lukka
//	 planeswalker."
//	 −2: Exile target creature you control, then reveal cards from the top of
//	 your library until you reveal a creature card with greater mana value. Put
//	 that card onto the battlefield and the rest on the bottom of your library
//	 in a random order.
//	 −7: Each creature you control deals damage equal to its power to each
//	 opponent."
//
// The +1 is Possibility Technician's conditional exile permission
// (game.CondControlSubtype, #3192): the three cards are exiled face up, and a
// permission to cast (not play) each CREATURE card lasts for as long as the
// card stays exiled, live only while its holder controls a Lukka (this one or
// another), dormant otherwise and live again when one returns. The filter
// leaves the other cards without a permission. The −2 notes the exiled
// creature's mana value as it left the battlefield (CR 608.2h, a token is 0),
// whether or not the exile happened, then Atla Palani's reveal-until puts the
// first creature card with a greater value onto the battlefield and the rest
// on the bottom in a random order. The −7 reads every creature as it
// resolves and deals all the damage in one damage instance (CR 120.2), each
// creature to each opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ca97394-4e8c-4698-bea1-554f9b10927b",
		Name:            "Lukka, Coppercoat Outcast",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Exile the top three cards of your library. Creature cards exiled this way gain \"You may cast this card from exile as long as you control a Lukka planeswalker.\"",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ExileTopWithPermission{
						From: item.Controller, GrantTo: item.Controller, N: 3,
						CastOnly:    true,
						WhileExiled: true,
						Condition:   &game.PermissionCondition{Key: game.CondControlSubtype, Subtype: ctx.TextSubtype("Lukka")},
						Filter:      game.PermissionFilter{CardInZone: true, CreatureOnly: true},
					}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Exile target creature you control, then reveal cards from the top of your library until you reveal a creature card with greater mana value. Put that card onto the battlefield and the rest on the bottom of your library in a random order.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  lukkaCoppercoatExile,
			},
			{
				Label:  "−7: Each creature you control deals damage equal to its power to each opponent.",
				Cost:   effects.LoyaltyCost(-7),
				Effect: lukkaCoppercoatDamage,
			},
		},
	})
}

// lukkaCoppercoatExile exiles the target and hands its mana value to the
// reveal that follows.
func lukkaCoppercoatExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.ExileTarget{Target: id, ThenTail: effects.WithN(ctx.ItemTail(lukkaCoppercoatReveal), c.ManaValue())}.Apply(ctx)
}

// lukkaCoppercoatReveal is "then reveal cards ... until you reveal a creature
// card with greater mana value ...". An on-disk key: never renamed, never
// reused.
var lukkaCoppercoatReveal = effects.CardTail("lukka-coppercoat-outcast/reveal-greater-creature", func(ctx *effects.Context, o game.TailOutcome) error {
	mv := o.Tail.N
	return effects.RevealUntilThenPutOntoBattlefield{
		Match:  func(c game.Card) bool { return c.IsCreature() && c.ManaValue() > mv },
		Reason: "Lukka, Coppercoat Outcast — revealed until a creature card with greater mana value",
	}.Apply(ctx)
})

// lukkaCoppercoatDamage is the -7: one damage instance, every creature's
// power at once, to every opponent.
func lukkaCoppercoatDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	g.RecomputeLayersIfStaleLocked()
	type hit struct {
		source uuid.UUID
		power  int
	}
	var hits []hit
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.CurrentPower() > 0 {
			hits = append(hits, hit{c.InstanceID, c.CurrentPower()})
		}
	}
	opponents := ctx.Opponents()
	return g.DamageInstanceForEffect(func() error {
		for _, h := range hits {
			for _, opp := range opponents {
				if err := (effects.DealDamage{Source: h.source, Target: opp, Amount: h.power}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
