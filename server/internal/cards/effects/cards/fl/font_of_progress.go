package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Font of Progress — Artifact {U} (Phyrexia: All Will Be One):
//
//	"This artifact enters with two oil counters on it.
//	 {3}, {T}: Target player mills X cards, where X is the number of oil
//	 counters on this artifact."
//
// Nothing is removed. X is counted as the ability resolves (CR 608.2h):
// live while the Font is still there, as it last existed if it has gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec1fcec2-ba7a-44d7-b251-22a7cd417493",
		Name:         "Font of Progress",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 2, "Font of Progress: enters with two oil counters on it")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target player mills X cards, where X is the number of oil counters on this artifact.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.TargetPlayerMills(effects.S82OneB10OilCounters(ctx))(g, item)
			},
		}},
	})
}
