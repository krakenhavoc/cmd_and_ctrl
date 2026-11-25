package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deafening Clarion — Sorcery {1}{R}{W} (Guilds of Ravnica, #5525):
//
//	"Choose one or both —
//	 • Deafening Clarion deals 3 damage to each creature.
//	 • Creatures you control gain lifelink until end of turn."
//
// With both chosen the modes run in printed order (CR 700.2), so the damage
// is dealt first and the lifelink goes to the creatures you control that are
// still on the battlefield as the second instruction resolves (CR 611.2c); it
// does not give the damage lifelink.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9bbebbd-9f2d-499c-adb1-8d3170c3fd0d",
		Name:         "Deafening Clarion",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeWithPurpose(effects.Mode("Deafening Clarion deals 3 damage to each creature."),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}}),
			effects.Mode("Creatures you control gain lifelink until end of turn."),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				if err := effects.DamageEachMatching(ctx, effects.Creature(), 3); err != nil {
					return err
				}
			}
			if ctx.HasMode(1) {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"lifelink"},
					Label:    "Deafening Clarion — lifelink until end of turn",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
