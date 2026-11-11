package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e802bf11-d3c7-4aad-8e85-99b39ff50d70",
		Name:         "Gourmand's Talent",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Activated: []ActivatedAbility{cmdr2S18FoodSacrificeAbility()},
			Text:      "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		}},
		Static: []game.StaticAbility{
			{
				Label:     "During your turn, artifacts you control are Foods in addition to their other types",
				Layer:     game.Layer4Type,
				AppliesTo: cmdr2S18ArtifactsYouControlOnYourTurn,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasType(c.Subtypes, "Food") {
						c.Subtypes = append(append([]string(nil), c.Subtypes...), "Food")
					}
				},
			},
			Labelled(GrantAbilities(cmdr2S18ArtifactsYouControlOnYourTurn, grantKey),
				"During your turn, artifacts you control have \"{2}, {T}, Sacrifice this artifact: You gain 3 life.\""),
		},
		Triggered: []game.TriggeredAbility{
			AtLevel(2, On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return socB28FirstLifeGainThisTurn(ev, source, g)
			}, "Gourmand's Talent — create a 3/3 green Raccoon creature token",
				Do(CreateToken{Template: TokenCard("3/3 green Raccoon"), N: 1}))),
			AtLevel(3, On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return socB28FirstLifeGainThisTurn(ev, source, g)
			}, "Gourmand's Talent — put a +1/+1 counter on each creature you control",
				Do(Instruction(putAPlusOneCounterOnEachCreatureYouControlStep)))),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{2}{G}")),
			LevelUp(3, ManaCost("{3}{G}")),
		},
	})
}
