package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Samite Pilgrim — Creature — Human Cleric {1}{W}, 1/1 (Planeshift, S82 OLD b642, #5615):
//
//	"Domain — {T}: Prevent the next X damage that would be dealt to target
//	 creature this turn, where X is the number of basic land types among lands
//	 you control."
//
// The shield is Samite Alchemist's charged PreventNextDamage (CR 615.7) pinned to
// the target. X is read as the ability resolves (CR 608.2c): the number of basic
// land types among the lands its controller controls then. A target that is no
// longer legal gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "542c5eaf-8b03-4406-a794-30c56b61d637",
		Name:         "Samite Pilgrim",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Domain — {T}: Prevent the next X damage that would be dealt to target creature this turn, where X is the number of basic land types among lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.PreventNextDamage{
					Target: id,
					Amount: effects.DomainCount(g, item.Controller),
					Label:  "Samite Pilgrim: prevent the next X damage",
				}.Apply(ctx)
			},
		}},
	})
}
