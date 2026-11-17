package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "23ea480c-99e3-4cbf-8d24-9ea92230f57b",
		Name:         "Wojek Apothecary",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to target creature and each other creature that shares a color with it this turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return eachRadiant(ctx, radianceSet(ctx, Creature()), func(c *Context, id uuid.UUID) error {
					return PreventNextDamage{Target: id, Amount: 1, Label: "Wojek Apothecary: prevent the next 1 damage"}.Apply(c)
				})
			},
		}},
	})
}
