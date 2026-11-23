package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grindstone — Artifact {1}:
//
//	"{3}, {T}: Target player mills two cards. If two cards that share a
//	 color were milled this way, repeat this process."
//
// Sphinx's Tutelage's loop without the nonland test (a card with no
// color shares none, CR 105.2c): each pass is one mill of two, so the
// passes are separate mill events, and the repeat ends when a pass mills
// fewer than two cards or the pair shares no color. The player is the one
// chosen on activation, re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97560adc-814e-450c-9ca7-f9364e910a0b",
		Name:         "Grindstone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target player mills two cards. If two cards that share a color were milled this way, repeat this process.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.GrindstoneMillPass(ctx, t.ID, 0)
					}
				}
				return nil
			},
		}},
	})
}
