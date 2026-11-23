package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hagra Constrictor — Creature — Snake {2}{B}, 0/0 (Zendikar Rising,
// #5508):
//
//	"This creature enters with two +1/+1 counters on it.
//	 Each creature you control with a +1/+1 counter on it has menace."
//
// The counters are a CR 614.1c entry replacement, so the 0/0 never meets
// a state-based check. The menace grant is Great Ugly-Looking Goblin's
// layer 6 grant over every creature its controller controls (itself
// included) that carries a +1/+1 counter, read off the counter map, so a
// creature gains menace with its first counter and loses it with the
// last.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3282c7fb-b43d-4dd6-b5a3-cc7d93e01f19",
		Name:         "Hagra Constrictor",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Hagra Constrictor — enters with two +1/+1 counters on it"),
		},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.B16CreaturesYouControl(target, g, source) && target.Counters[game.CounterPlusOne] > 0
			}, "menace"),
		},
	})
}
