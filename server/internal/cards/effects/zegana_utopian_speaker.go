package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zegana, Utopian Speaker — Legendary Creature — Merfolk Wizard {2}{G}{U},
// 4/4 (Lost Caverns of Ixalan Commander, #5368):
//
//	"When Zegana enters, if you control another creature with a +1/+1
//	 counter on it, draw a card.
//	 {4}{G}{U}: Adapt 4.
//	 Each creature you control with a +1/+1 counter on it has trample."
//
// The draw is an intervening "if" (CR 603.4), Mine Raider's shape; Zegana
// herself never counts toward it. Adapt checks for counters as it resolves
// (CR 701.46a). The trample grant reads the counter live and includes
// Zegana (Duskshell Crawler's grant).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04260930-ff38-4b35-9ec4-e89196a4f2c7",
		Name:         "Zegana, Utopian Speaker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16GrantKeywords(b42CreatureYouControlWithAPlusOneCounter, "trample")},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control another creature with a +1/+1 counter on it",
				Test: func(ctx *Context) bool {
					return cmdr2B23AnotherCreatureWithPlusOne(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)
				},
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && cmdr2B23AnotherCreatureWithPlusOne(g, source.Controller, source.InstanceID)
			}, "Zegana — draw a card", Do(DrawCards{N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}{U}: Adapt 4",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{4}{G}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 4) },
		}},
	})
}
