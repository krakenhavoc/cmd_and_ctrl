package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heartless Act — Instant {1}{B}:
//
//	"Choose one —
//	 • Destroy target creature with no counters on it.
//	 • Remove up to three counters from target creature."
//
// The mode and its target are chosen as the spell is cast (CR 700.2a).
// The first mode's "no counters" is rechecked on resolution: a creature
// that gained a counter in response is an illegal target and the spell
// does nothing, with no switching to the other mode (CR 608.2b, the
// ruling). The second mode targets any creature and, as it resolves,
// asks the caster how many of each kind of counter to remove, up to
// three in all, whoever controls the creature (Glissa Sunslayer's
// clause, CR 608.2c, the ruling).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3641d572-8335-4804-88ad-edf4dc67a8e4",
		Name:         "Heartless Act",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target creature with no counters on it.",
				effects.TargetCreature("target creature with no counters on it", effects.B10NoCounters()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Remove up to three counters from target creature.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GlissaSunslayerRemoves.Choose(ctx, t.ID, nil, 3, "Heartless Act — remove up to three counters")
				}),
		),
	})
}
