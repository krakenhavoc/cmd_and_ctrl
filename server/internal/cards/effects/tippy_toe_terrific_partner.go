package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tippy-Toe, Terrific Partner — Legendary Creature — Squirrel Hero {3}{G},
// 2/4 (Marvel Super Heroes Commander, #4455):
//
//	"If you would create one or more tokens, instead create those tokens plus
//	 an additional Food token.
//	 At the beginning of your end step, if you gained life this turn, draw a
//	 card."
//
// Xorn's CR 614 token-creation replacement without the Treasure condition:
// it applies to any creation under Tippy-Toe's controller and appends one
// Food group, once per instruction, leaving the instruction's own tokens
// alone. Two Tippy-Toes each add one, and a creation of Food alone gets one
// more Food. The end-step draw is Ocelot Pride's intervening "if"
// (CR 603.4), checked as the step begins and again as the trigger resolves
// against the life gained this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a1bce89-0c11-4d8d-aacb-c0a5a5effffc",
		Name:         "Tippy-Toe, Terrific Partner",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens && ev.TokenCount() > 0 &&
					ev.TokenController == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{Template: ruleDefinedToken(FoodToken()), Count: 1})
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Tippy-Toe: an additional Food token",
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you gained life this turn",
				Test: func(ctx *Context) bool { return b15LifeGainedThisTurn(ctx.Game, ctx.Controller()) > 0 },
			}, On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24YourEndStepAndYouGainedLifeThisTurn(ev, source, g)
			}, "Tippy-Toe — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
