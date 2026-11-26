package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nivix Barrier — Creature — Illusion Wall {3}{U}, 0/4 (Magic Origins, S82 older
// sets):
//
//	"Flash
//	 Defender
//	 When this creature enters, target attacking creature gets -4/-0 until end
//	 of turn."
//
// The shrink is a layer 7c modifier on the chosen attacker, set as the trigger
// resolves (CR 611.2c). The target must be attacking as the trigger is put on the
// stack and again as it resolves (CR 608.2b), so a Barrier flashed in outside
// combat has no target and its trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa4cb5fb-e81c-410f-8986-645671b0a1d3",
		Name:            "Nivix Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "defender"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Nivix Barrier — target attacking creature gets -4/-0 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								return effects.BoostUntilEOT{Target: t.ID, Power: -4, Label: "Nivix Barrier — -4/-0"}.Apply(ctx)
							}
						}
						return nil
					}),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())),
		},
	})
}
