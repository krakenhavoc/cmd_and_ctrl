package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret's Simulacrum — Artifact Creature — Golem {3}, 2/3 (Aether Revolt,
// S82 older sets, #5536):
//
//	"{T}: Target opponent loses 1 life. If you control a Tezzeret
//	 planeswalker, that player loses 3 life instead."
//
// The Tezzeret planeswalker is looked for as the ability resolves (CR 608.2c,
// 608.2h), by its subtype through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd886509-7455-4c8f-976f-0f5fafcb97be",
		Name:         "Tezzeret's Simulacrum",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target opponent loses 1 life. If you control a Tezzeret planeswalker, that player loses 3 life instead.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := 1
				if effects.CountControlled(g, ctx.Controller(), func(c game.Card) bool {
					return c.IsPlaneswalker() && c.HasSubtype("Tezzeret")
				}) > 0 {
					n = 3
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
					}
				}
				return nil
			},
		}},
	})
}
