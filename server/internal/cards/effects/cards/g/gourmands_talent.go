package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gourmand's Talent — Enchantment — Class {G} (Bloomburrow Commander,
// #5368):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 During your turn, artifacts you control are Foods in addition to their
//	 other types and have '{2}, {T}, Sacrifice this artifact: You gain 3
//	 life.'
//	 {2}{G}: Level 2
//	 Whenever you gain life for the first time each turn, create a 3/3
//	 green Raccoon creature token.
//	 {3}{G}: Level 3
//	 Whenever you gain life for the first time each turn, put a +1/+1
//	 counter on each creature you control."
//
// Level 1's static is two layered effects over the same set (layer 4 adds
// the Food subtype, layer 6 grants the Food ability by key), read on your
// turn only. The level 2 and 3 lines are level-gated triggers off your
// first life gain of the turn, as read from the turn's event log.
func init() {
	const grantKey = "cmdr2-s18/gourmands-talent-food"
	effects.Register(effects.Spec{
		OracleID:     "e802bf11-d3c7-4aad-8e85-99b39ff50d70",
		Name:         "Gourmand's Talent",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Activated: []effects.ActivatedAbility{effects.Cmdr2S18FoodSacrificeAbility()},
			Text:      "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		}},
		Static: []game.StaticAbility{
			{
				Label:     "During your turn, artifacts you control are Foods in addition to their other types",
				Layer:     game.Layer4Type,
				AppliesTo: effects.Cmdr2S18ArtifactsYouControlOnYourTurn,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Subtypes, "Food") {
						c.Subtypes = append(append([]string(nil), c.Subtypes...), "Food")
					}
				},
			},
			effects.Labelled(effects.GrantAbilities(effects.Cmdr2S18ArtifactsYouControlOnYourTurn, grantKey),
				"During your turn, artifacts you control have \"{2}, {T}, Sacrifice this artifact: You gain 3 life.\""),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtLevel(2, effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.SocB28FirstLifeGainThisTurn(ev, source, g)
			}, "Gourmand's Talent — create a 3/3 green Raccoon creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Raccoon"), N: 1}))),
			effects.AtLevel(3, effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.SocB28FirstLifeGainThisTurn(ev, source, g)
			}, "Gourmand's Talent — put a +1/+1 counter on each creature you control",
				effects.Do(effects.Instruction(effects.PutAPlusOneCounterOnEachCreatureYouControlStep)))),
		},
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{2}{G}")),
			effects.LevelUp(3, effects.ManaCost("{3}{G}")),
		},
	})
}
