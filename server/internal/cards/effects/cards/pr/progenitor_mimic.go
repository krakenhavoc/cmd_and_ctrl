package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Progenitor Mimic — Creature — Shapeshifter {4}{G}{U}, 0/0 (Commander 2016,
// S82 older sets, #5537):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it has "At the beginning of your upkeep, if this
//	 creature isn't a token, create a token that's a copy of this creature.""
//
// The copy effect is an enters-as-a-copy replacement (CR 614.1c, 707.9b)
// whose exception is a granted bundle: the copy has the upkeep ability, and
// because the ability is part of what is copied, the token it makes has it
// too, but is a token and so stops the chain (the intervening "if", CR
// 603.4). Declining to copy leaves a 0/0 that dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88929ea9-900f-4dbb-b16c-cf3bad4e410c",
		Name:         "Progenitor Mimic",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  progenitorMimicUpkeepGrant,
			Text: "At the beginning of your upkeep, if this creature isn't a token, create a token that's a copy of this creature.",
			Triggered: []game.TriggeredAbility{
				effects.InterveningIf(effects.CondBoard{
					Name: "this creature isn't a token",
					Test: func(ctx *effects.Context) bool {
						c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
						return ok && !c.IsToken()
					},
				}, effects.AtYourUpkeep("Progenitor Mimic — create a token that's a copy of this creature",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.SourceIsStillThisPermanent(g, item) {
							return nil
						}
						return effects.CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(effects.NewContext(g, item))
					})),
			},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Progenitor Mimic", effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(progenitorMimicUpkeepGrant)
				}),
		},
	})
}

// progenitorMimicUpkeepGrant is the granted bundle's key, an on-disk
// identity: never renamed.
const progenitorMimicUpkeepGrant = "progenitor-mimic/upkeep-copy"
