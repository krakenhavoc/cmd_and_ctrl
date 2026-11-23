package cardsgr

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grist, Voracious Larva // Grist, the Plague Swarm — Modern Horizons 3
// (#5948), a transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Insect {G}, 1/2:
//
//	"Deathtouch
//	 Whenever Grist or another creature you control enters, if it entered
//	 from your graveyard or you cast it from your graveyard, you may pay {G}.
//	 If you do, exile Grist, then return it to the battlefield transformed
//	 under its owner's control."
//
// Back face, Legendary Planeswalker — Grist, loyalty 3:
//
//	"+1: Create a 1/1 black and green Insect creature token, then mill two
//	     cards. Put a deathtouch counter on the token if a black card was
//	     milled this way.
//	 −2: Destroy target artifact or enchantment.
//	 −6: For each creature card in your graveyard, create a token that's a
//	     copy of it, except it's a 1/1 black and green Insect."
//
// The back face registers under "<oracle_id>#1" (game.CatalogKey). The
// trigger's "if" (CR 603.4) is Archfiend's Vessel's, for any creature you
// control: the entry came from your graveyard, or you cast the creature from
// it. The payment is paid as the trigger resolves (CR 118.12, MayPay). The
// +1's three sentences are a token creation whose tail mills, whose tail puts
// the counter. The −6 is one creation of every copy together (CR 608.2f), each
// copying its card as it sits in the graveyard (a Grist that died to the cost
// is the front face, CR 712.8a), the exception setting colors, creature types
// and 1/1 as copiable values (CR 707.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        gristVoraciousLarvaOracle,
		Name:            "Grist, Voracious Larva",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "it entered from your graveyard or you cast it from your graveyard",
				Test: func(ctx *effects.Context) bool {
					src, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
					if !ok {
						return gristEnteredFromGraveyardEvent(ctx.Trigger().Event, ctx.Controller())
					}
					return gristCreatureReturned(ctx.Trigger().Event, &src, game.Characteristic{}, ctx.Game)
				},
			}, effects.On(game.EventETB, gristCreatureReturned,
				"Grist, Voracious Larva — you may pay {G}; if you do, exile Grist, then return it transformed",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{G}",
						Question: "Pay {G}: exile Grist, then return it to the battlefield transformed?",
						Then:     gristPaid,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        gristVoraciousLarvaOracle + "#1",
		Name:            "Grist, the Plague Swarm",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 black and green Insect creature token, then mill two cards. Put a deathtouch counter on the token if a black card was milled this way.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups: []game.TokenGroup{{
							Template: ctx.TextToken(effects.TokenCard("1/1 black and green Insect")),
							Count:    1,
						}},
					}, ctx.ItemTail(gristInsectThenMill))
				},
			},
			{
				Label:   "−2: Destroy target artifact or enchantment.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
				},
			},
			{
				Label: "−6: For each creature card in your graveyard, create a token that's a copy of it, except it's a 1/1 black and green Insect.",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					p := g.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Graveyard == nil {
						return nil
					}
					var groups []game.TokenGroup
					for _, c := range p.Graveyard.Cards {
						if !c.IsCreature() {
							continue
						}
						// A card in a graveyard is copied as the card it is (CR 707.2),
						// so a transformed planeswalker that died is its front face.
						tmpl := effects.TokenCopyTemplateFromCard(c)
						gristInsectCopyException(&tmpl)
						groups = append(groups, game.TokenGroup{Template: effects.CopiedToken(tmpl), Count: 1})
					}
					if len(groups) == 0 {
						return nil
					}
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: ctx.Controller(), Source: ctx.Source(), Groups: groups, Together: true,
					}, game.KeyedTail{})
				},
			},
		},
	})
}

// gristVoraciousLarvaOracle is shared by both faces.
const gristVoraciousLarvaOracle = "196f8f13-d4df-4740-8746-8fed3bc71840"

// gristEnteredFromGraveyardEvent is the entry-event half of the condition, for
// a creature that is no longer anywhere it can be looked up.
func gristEnteredFromGraveyardEvent(ev game.Event, you uuid.UUID) bool {
	return ev.Kind == game.EventETB && ev.EnteredFrom == game.ZoneGraveyard && ev.EnteredFromOwner == you
}

// gristCreatureReturned is "Grist or another creature you control enters, if it
// entered from your graveyard or you cast it from your graveyard".
func gristCreatureReturned(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || c.Controller != source.Controller {
		return false
	}
	if gristEnteredFromGraveyardEvent(ev, source.Controller) {
		return true
	}
	cast, known := c.CastByItsController()
	return known && cast && c.Provenance.FromZone == game.ZoneGraveyard
}

// gristPaid is the paid branch: exile Grist, then return it transformed. It is
// still the permanent the trigger came from only while it is that object
// (ExileAndReturnTransformed's CR 400.7 guard).
var gristPaid = effects.PayUnlessBranches("pay-unless/grist-voracious-larva-transform",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ExileAndReturnTransformed{Target: ctx.Source()}.Apply(ctx)
	}, nil)

// gristInsectThenMill is the sentence after the token: mill two cards, the
// token riding the mill's tail to take the counter. An on-disk key.
var gristInsectThenMill = effects.CardTail("grist-the-plague-swarm/mill-two", func(ctx *effects.Context, o game.TailOutcome) error {
	var token uuid.UUID
	if len(o.IDs) > 0 {
		token = o.IDs[0]
	}
	tail := ctx.ItemTail(gristMilledBlack)
	if token != uuid.Nil {
		tail.Carry = []uuid.UUID{token}
	}
	return effects.MillToZone{N: 2, ThenTail: tail}.Apply(ctx)
})

// gristMilledBlack puts the deathtouch counter on the token when a black card
// was among the cards milled this way. An on-disk key.
var gristMilledBlack = effects.CardTail("grist-the-plague-swarm/deathtouch-counter", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, col := range c.EffectiveColors() {
			if col == "B" {
				return effects.AddCounter{Target: o.Tail.Carry[0], Kind: game.CounterDeathtouch, N: 1}.Apply(ctx)
			}
		}
	}
	return nil
})

// gristInsectCopyException is "except it's a 1/1 black and green Insect": the
// copiable colors, creature types and power and toughness (CR 707.9b). Its
// other card types stay.
func gristInsectCopyException(t *game.Card) {
	t.Power, t.Toughness = 1, 1
	t.VariableToughness = false
	t.PrintedPTKnown = true
	t.Colors = []string{"B", "G"}
	if i := strings.Index(t.TypeLine, " — "); i >= 0 {
		t.TypeLine = t.TypeLine[:i]
	}
	t.TypeLine += " — Insect"
}
