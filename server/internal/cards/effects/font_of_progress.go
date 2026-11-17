package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ec1fcec2-ba7a-44d7-b251-22a7cd417493",
		Name:         "Font of Progress",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 2, "Font of Progress: enters with two oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Target player mills X cards, where X is the number of oil counters on this artifact.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return TargetPlayerMills(s82OneB10OilCounters(ctx))(g, item)
			},
		}},
	})
}
