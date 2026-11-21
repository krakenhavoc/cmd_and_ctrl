package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Wolf — Creature — Wolf {3}{R}, 4/3 (Innistrad: Crimson Vow,
// #5495):
//
//	"{1}{R}: This creature gains first strike until end of turn.
//	 Activate only as a sorcery."
//
// Cobalt Golem's self grant at {1}{R}, held to sorcery speed
// (CR 307.5): the activation is refused with the stack non-empty, in a
// combat step or on another player's turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d50f1ff7-d2bc-4e36-876d-77648f447f5b",
		Name:         "Lightning Wolf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{R}: This creature gains first strike until end of turn. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{1}{R}"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"},
					Label: "Lightning Wolf — first strike"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
