package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Passwall Adept — Creature — Human Wizard {1}{U}, 1/3 (Ravnica: Clue Edition, #5929):
//
//	"{2}{U}: Target creature can't be blocked this turn."
//
// Access Tunnel's RestrictUntilEOT(CantBeBlocked) with no tap and no power cap; a target that left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a2bc266-4076-4923-b201-5898b18cbad9",
		Name:         "Passwall Adept",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Passwall Adept — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
