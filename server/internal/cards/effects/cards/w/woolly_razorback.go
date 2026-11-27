package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woolly Razorback — Creature — Boar Beast {2}{W}{W}, 7/7 (Coldsnap, #5597):
//
//	"This creature enters with three ice counters on it.
//	 As long as this creature has an ice counter on it, prevent all combat
//	 damage it would deal and it has defender.
//	 Whenever this creature blocks, remove an ice counter from it."
//
// Three entry counters (a self-replacement, CR 614.1c); the keyword is a
// layer 6 static read live off the counters, the prevention a standing CR 615
// replacement of this creature's combat damage, read as the damage would be
// dealt. The block trigger fires once however many attackers it blocks
// (CR 509.3a).
//
// No simplification.
func init() {
	hasIce := func(src *game.Card) bool { return src.Counters[effects.S82OldB565IceCounter] > 0 }
	effects.Register(effects.Spec{
		OracleID:     "c5d27469-96e1-44ec-a887-2fc26f9e10d3",
		Name:         "Woolly Razorback",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(effects.S82OldB565IceCounter, 3, "Enters with three ice counters on it."),
			effects.S82OldB565PreventDamageDealtBy("Woolly Razorback — prevent all combat damage it would deal", true,
				func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) bool { return hasIce(src) }),
		},
		Static: []game.StaticAbility{{
			Label: "As long as this creature has an ice counter on it, it has defender.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && hasIce(target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "defender")
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Woolly Razorback — remove an ice counter from it",
				effects.RemoveACounterFromThis(effects.S82OldB565IceCounter)),
		},
	})
}
