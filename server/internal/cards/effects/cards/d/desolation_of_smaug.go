package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desolation of Smaug — Sorcery {2}{R}{R} (The Hobbit):
//
//	"Desolation of Smaug deals 3 damage to each non-Dragon creature.
//	 Add four mana in any combination of colors. Spend this mana only to
//	 cast Dragon spells."
//
// The first sentence is Caught in the Crossfire's untargeted sweep
// (damageEachMatching over creatures that are not Dragons, the Dragon test
// reading effective subtypes), declared a partial creature sweep for the
// bot. The second is Conversion Apparatus's four independent any-colour
// picks (AnyCombinationOfColors), restricted to casting a spell with the
// Dragon subtype. The mana is added after the damage has been dealt, in
// printed order, and a prompt the damage asks (CR 608.2c) delays it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9634afc-4a5b-4cf6-b63d-0ff9909dd5a7",
		Name:         "Desolation of Smaug",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.Not(effects.Subtype("Dragon"))), 3)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.AddMana{
					Produced:     effects.AnyCombinationOfColors(4),
					Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Dragon")},
				}.Apply(ctx)
			}),
		),
	})
}
