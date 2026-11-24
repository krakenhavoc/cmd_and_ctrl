package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Symmetry Sage — Creature — Human Wizard {U}, 0/2 (Strixhaven: School of
// Mages, #5502):
//
//	"Flying
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 target creature you control has base power 2 until end of turn."
//
// Flying is the printed keyword. The magecraft is a layer 7b base-power
// set (CR 613.4b) on the chosen creature, ending in cleanup, so a counter
// or a +1/+0 pump still adds to the new base.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16cb916d-bb19-42ec-9a01-969e9d067b40",
		Name:            "Symmetry Sage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Magecraft("Symmetry Sage — target creature you control has base power 2 until end of turn (magecraft)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								return effects.UntilEndOfTurn(ctx, t.ID, nil, "Symmetry Sage — base power 2", game.SetBasePowerMod(2))
							}
						}
						return nil
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
