package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worldwalker Helm — Artifact {2}{U} (The Big Score, #4468):
//
//	"If you would create one or more artifact tokens, instead create those
//	 tokens plus an additional Map token.
//	 {1}{U}, {T}: Create a token that's a copy of target artifact token you
//	 control."
//
// Xorn's shape (a CR 614 token-creation replacement that appends a group
// once per instruction), keyed on an artifact among the tokens being
// created. The Map is part of the same event, so it is not itself replaced
// again (CR 614.5), and it carries none of the other tokens' abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9573c85a-e574-4b4e-aae0-2165c45fd27e",
		Name:         "Worldwalker Helm",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller &&
					ev.TokenTemplatesMatch(func(t game.Card) bool { return t.IsArtifact() })
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{Template: ruleDefinedToken(MapToken()), Count: 1})
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Worldwalker Helm: an additional Map",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Create a token that's a copy of target artifact token you control.",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: b35ArtifactTokenYouControl(),
			Effect:  trekB39TokenCopyOfTheTarget,
		}},
	})
}
