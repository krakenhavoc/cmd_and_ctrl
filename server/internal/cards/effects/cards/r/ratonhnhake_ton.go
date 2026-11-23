package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ratonhnhaké꞉ton — Legendary Creature — Human Assassin {W}{U}{B}, 3/3
// (Assassin's Creed, #5948):
//
//	"As long as Ratonhnhaké꞉ton hasn't dealt damage yet, it has hexproof and
//	 can't be blocked.
//	 Whenever Ratonhnhaké꞉ton deals combat damage to a player, create a 1/1
//	 black Assassin creature token with menace. When you do, return target
//	 Equipment card from your graveyard to the battlefield, then attach it to
//	 that token."
//
// The first sentence is Karakyk Guardian's gate (game.HasntDealtDamage: the
// object-history mark of damage of any kind, kept until it leaves the
// battlefield), on two layer 6 statics, hexproof and "can't be blocked". The
// second is Forum Filibuster's token-then-reflexive shape (CR 603.12): the
// token is created first, and only if it was is the reflexive trigger put on
// the stack, whose Equipment target (a card in your graveyard, chosen as it
// goes on the stack) enters attached to that token (ReturnFromGraveyardTogether).
// With no Equipment in the graveyard the reflexive trigger is removed
// (CR 603.3d) and the token stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a37f8381-83a4-4b6b-8176-c90f6a61da74",
		Name:         "Ratonhnhaké꞉ton",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:      "As long as this creature hasn't dealt damage yet, it has hexproof.",
				Layer:      game.Layer6Ability,
				ActiveWhen: game.HasntDealtDamage(),
				AppliesTo:  effects.SelfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					effects.AppendKeywordsTo(c, []string{"hexproof"})
				},
			},
			{
				Label:      "As long as this creature hasn't dealt damage yet, it can't be blocked.",
				Layer:      game.Layer6Ability,
				ActiveWhen: game.HasntDealtDamage(),
				AppliesTo:  effects.SelfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Restrictions |= game.CantBeBlocked
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Ratonhnhaké꞉ton — create a 1/1 black Assassin creature token with menace; when you do, return target Equipment card from your graveyard to the battlefield, then attach it to that token",
				ratonhnhaketonToken),
		},
	})
}

func ratonhnhaketonToken(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(effects.TokenCard("1/1 black Assassin with menace")),
			Count:    1,
		}},
	}, ctx.ItemTail(ratonhnhaketonTokenCreated))
}

// ratonhnhaketonTokenCreated is "when you do". An on-disk key.
var ratonhnhaketonTokenCreated = effects.CardTail("ratonhnhaketon/token-created", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label: "Ratonhnhaké꞉ton — return target Equipment card from your graveyard to the battlefield, then attach it to that token",
		Cards: o.IDs[:1],
		Body:  ratonhnhaketonReturnBody,
	}.Apply(ctx)
})

// ratonhnhaketonReturnBody is the reflexive trigger's registration: the token
// is the payload, the Equipment card the trigger's own target. An on-disk key.
var ratonhnhaketonReturnBody = game.ReflexiveBody("ratonhnhaketon/return-equipment", effects.SimpleBody(ratonhnhaketonReturn),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("target Equipment card from your graveyard", effects.YouOwn(), effects.Subtype("Equipment"))
	}))

// ratonhnhaketonReturn returns the target attached to the token, if the token
// is still on the battlefield.
func ratonhnhaketonReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	tokens := ctx.PayloadCards()
	targets := ctx.LegalTargets()
	if len(tokens) == 0 || len(targets) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{targets[0].ID},
		Controller: item.Controller,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: tokens[0]},
	}.Apply(ctx)
}
