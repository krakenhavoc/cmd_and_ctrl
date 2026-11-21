package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Magic — Instant {R} (#3777):
//
//	"Tiered (Choose one additional cost.)
//	 • Fire — {0} — Fire Magic deals 1 damage to each creature.
//	 • Fira — {2} — Fire Magic deals 2 damage to each creature.
//	 • Firaga — {5} — Fire Magic deals 3 damage to each creature."
//
// Tiered proof card (CR 702.183a): untargeted bullets, so the tier
// prompt is the only question the cast asks. Each bullet is a sweep.
//
// No simplification.
func init() {
	bullet := func(label, cost string, n int) game.ModeOption {
		return effects.ModeWithPurpose(effects.SpreeModeDoing(label, cost, nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DamageEachMatching(ctx, effects.Creature(), n)
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: n}})
	}
	effects.Register(effects.Spec{
		OracleID:     "c87db803-3490-41bc-af7a-78f99bcd77c8",
		Name:         "Fire Magic",
		Completeness: effects.CompletenessFull,
		Modes: effects.Tiered(
			bullet("Fire — Fire Magic deals 1 damage to each creature.", "{0}", 1),
			bullet("Fira — Fire Magic deals 2 damage to each creature.", "{2}", 2),
			bullet("Firaga — Fire Magic deals 3 damage to each creature.", "{5}", 3),
		),
	})
}
