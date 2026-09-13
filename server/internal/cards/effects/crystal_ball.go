package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Ball — Artifact {3}:
//
//	"{1}, {T}: Scry 2."
//
// The repeatable scry. Nothing clever: it is the Scry primitive
// behind an ordinary mana-plus-tap activation, and it is in the
// catalog because a draw deck wants an engine that smooths every
// turn rather than a one-shot that smooths once.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "bd85fe4d-1d62-416f-ac2d-e287911c84e3",
		Name:     "Crystal Ball",
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Scry 2.",
			Cost:  game.AbilityCost{Mana: "{1}", Tap: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
