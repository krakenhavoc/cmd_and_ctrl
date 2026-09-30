package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stridehangar Automaton — Artifact Creature — Construct {3}, 1/4
// (slice 296-m):
//
//	"Thopters you control get +1/+1.
//	 If one or more artifact tokens would be created under your
//	 control, those tokens plus an additional 1/1 colorless Thopter
//	 artifact creature token with flying are created instead."
//
// The anthem is a plain Layer 7c static over "Thopters you control" —
// a fixed subtype, not a chosen one, so it is a one-off AppliesTo
// rather than tribal.go's TribeFilter{Chosen: true}.
//
// The token half is the CR 701.7b replacement window
// (game.RepEventCreateTokens), the same one Doubling Season and
// Parallel Lives multiply: TokenTemplatesMatch narrows to an
// instruction that is actually making an artifact token, under this
// controller, and Replace APPENDS a group rather than multiplying the
// existing ones — "those tokens PLUS an additional […] token", not
// "twice that many".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9070c98b-fd01-4eeb-a4ec-fc464946c7c0",
		Name:         "Stridehangar Automaton",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype("Thopter")
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventTokenCreated},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventCreateTokens || ev.TokenController != src.Controller {
						return false
					}
					return ev.TokenTemplatesMatch(func(c game.Card) bool { return c.IsArtifact() })
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{
						Template: TokenCard("1/1 colorless Thopter artifact with flying"),
						Count:    1,
					})
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Stridehangar Automaton: an additional Thopter",
			},
		},
	})
}
