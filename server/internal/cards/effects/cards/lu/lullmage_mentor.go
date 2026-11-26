package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lullmage Mentor — Creature — Merfolk Wizard {1}{U}{U}, 2/2 (Zendikar, S82 OLD
// b496):
//
//	"Whenever a spell or ability you control counters a spell, you may create a 1/1
//	 blue Merfolk creature token.
//	 Tap seven untapped Merfolk you control: Counter target spell."
//
// The trigger is Baral, Chief of Compliance's read of a counter event caused by
// an effect the Mentor's controller controls (a spell countered by the rules, as
// when its targets are gone, is not one). The ability's cost taps seven untapped
// Merfolk, the Mentor itself among them if it is one (it is), with no {T} symbol
// involved, so summoning sickness does not stop it (CR 302.6 only governs {T}).
// The new token's own tap pays too, as a Merfolk.
//
// No simplification.
func init() {
	const tapSeven = "seven untapped Merfolk you control"
	effects.Register(effects.Spec{
		OracleID:     "752566b4-d3f5-4f10-b6f5-7185809f7a9d",
		Name:         "Lullmage Mentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCounterSpell, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCounterSpell && ev.Target == ev.CardID &&
					ev.Cause == game.MoveCauseEffect && ev.CauseController == source.Controller
			}, "Lullmage Mentor — create a 1/1 blue Merfolk token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk"), N: 1})),
				"Lullmage Mentor — create a 1/1 blue Merfolk creature token?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Tap seven untapped Merfolk you control: Counter target spell.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  7,
				Filter: effects.TargetPermanent(tapSeven, effects.HasSubtype("Merfolk")),
				Label:  tapSeven,
			}},
			Targets: effects.TargetSpell("target spell"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CounterTheTargetSpell(item, effects.NewContext(g, item))
			},
		}},
	})
}
