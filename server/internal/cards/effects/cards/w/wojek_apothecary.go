package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wojek Apothecary — Creature — Human Cleric {2}{W}{W}, 1/1:
//
//	"Radiance — {T}: Prevent the next 1 damage that would be dealt to
//	 target creature and each other creature that shares a color with it
//	 this turn."
//
// The creatures are read as the ability resolves (radianceSet, #5736),
// whoever controls them, the Apothecary itself included when it shares a
// colour with the target, and each gets its own one-point shield
// (PreventNextDamage, CR 615.7) pinned to it as an object and gone at
// cleanup. A {T} ability, so it waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23ea480c-99e3-4cbf-8d24-9ea92230f57b",
		Name:         "Wojek Apothecary",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to target creature and each other creature that shares a color with it this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.EachRadiant(ctx, effects.RadianceSet(ctx, effects.Creature()), func(c *effects.Context, id uuid.UUID) error {
					return effects.PreventNextDamage{Target: id, Amount: 1, Label: "Wojek Apothecary: prevent the next 1 damage"}.Apply(c)
				})
			},
		}},
	})
}
