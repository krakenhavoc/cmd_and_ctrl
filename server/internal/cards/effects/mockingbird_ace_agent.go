package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mockingbird, Ace Agent — Legendary Creature — Human Spy Hero {3}{W},
// 2/2 (Marvel Super Heroes, #4455):
//
//	"Double strike
//	 Whenever you cast a spell that targets a creature you control, put
//	 a +1/+1 counter on Mockingbird."
//
// Read at the cast (CR 601.2i) from the targets the spell announced,
// Danitha's shape: a creature on the battlefield you control. A spell
// with two such targets still triggers once. Mockingbird herself is a
// creature you control, so a spell that targets her counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1fd00656-83ba-4feb-b089-224007936b99",
		Name:            "Mockingbird, Ace Agent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				return rfCreatureACastSpellTargetsAny(g, ev.CardID, func(t game.TargetRef) bool {
					c, ok := rfCreatureATargetIsCreatureOnBattlefield(g, t)
					return ok && c.Controller == source.Controller
				})
			}, "Mockingbird — put a +1/+1 counter on Mockingbird",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
