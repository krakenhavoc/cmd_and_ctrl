package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alert Heedbonder — Creature — Human Scout {1}{G/W}{G/W}, 2/4 (Ikoria: Lair
// of Behemoths, S82 OLD batch 120, #5512):
//
//	"Vigilance
//	 At the beginning of your end step, you gain 1 life for each creature you
//	 control with vigilance."
//
// The count is taken as the trigger resolves, through the layered view, so a
// creature that was given vigilance counts and the Heedbonder itself does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "683a06b6-af77-440f-81b9-8958597a7324",
		Name:            "Alert Heedbonder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Alert Heedbonder — you gain 1 life for each creature you control with vigilance",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && game.HasKeyword(&c, "vigilance") {
							n++
						}
					}
					return effects.GainLife{Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
