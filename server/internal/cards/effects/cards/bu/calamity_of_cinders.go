package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Calamity of Cinders — Sorcery {5}{R}{R}:
//
//	"Convoke (Your creatures can help cast this spell. Each creature you tap
//	 while casting this spell pays for {1} or one mana of that creature's
//	 color.)
//	 Calamity of Cinders deals 6 damage to each untapped creature."
//
// Convoke is the engine's cost-time tap. Tapping creatures for it is the
// reason the wipe hits the fewest creatures the caster paid with: untapped
// is judged as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbcf0c1c-29a7-4e83-8b06-869f83be49cd",
		Name:         "Calamity of Cinders",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.Untapped()), 6)
		},
	})
}
