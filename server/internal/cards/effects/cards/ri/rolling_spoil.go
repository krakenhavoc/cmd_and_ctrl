package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rolling Spoil — Sorcery {2}{G}{G} (Ravnica: City of Guilds, S82 OLD batch
// 581, #5600):
//
//	"Destroy target land. If {B} was spent to cast this spell, all creatures
//	 get -1/-1 until end of turn."
//
// The second sentence reads the mana that paid for this spell (ReadsOwnPayment,
// colour B), as it resolves, and follows the destroy as a later step of one
// walk (CR 608.2c). With the land an illegal target the spell does nothing
// (CR 608.2b), the -1/-1 included. Creatures that enter later this turn are not
// affected (CR 611.2c): the boost is applied to the creatures on the
// battlefield as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "9abe09fb-17ab-41e8-9d89-1467ed62f87b",
		Name:           "Rolling Spoil",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"B"},
		// A wipe in the audit's eyes: -1/-1 to every creature, but only when {B}
		// was spent.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}},
		Targets: effects.TargetPermanent("target land", effects.Land()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.KineticOozeDestroy),
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.ManaSpentOfColor("B") == 0 {
					return nil
				}
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -1, Toughness: -1, Label: "Rolling Spoil — -1/-1 until end of turn"}.Apply(ctx)
			}),
		),
	})
}
