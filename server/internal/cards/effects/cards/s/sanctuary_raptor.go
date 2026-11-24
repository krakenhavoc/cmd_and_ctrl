package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctuary Raptor — Artifact Creature — Bird {3}, 2/1 (Modern Horizons 2,
// #5501):
//
//	"Flying
//	 Whenever this creature attacks, if you control three or more tokens,
//	 this creature gets +2/+0 and gains first strike until end of turn."
//
// An intervening "if" (CR 603.4): three tokens as it attacks, and again as
// the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6af53351-ad10-4774-9b5c-7b58bea65bbc",
		Name:            "Sanctuary Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82OldB39ControlsThreeOrMoreTokens(), effects.On(game.EventAttack, effects.ThisAttacked,
				"Sanctuary Raptor — +2/+0 and first strike until end of turn (three or more tokens)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Sanctuary Raptor — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"},
						Label: "Sanctuary Raptor — first strike"}.Apply(ctx)
				})),
		},
	})
}
