package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inscription of Abundance — {1}{G} Instant:
//
//	"Kicker {2}{G}
//	 Choose one. If this spell was kicked, choose any number instead.
//	 • Put two +1/+1 counters on target creature.
//	 • Target player gains X life, where X is the greatest power among
//	   creatures they control.
//	 • Target creature you control fights target creature you don't
//	   control."
//
// Inscription of Ruin's count (#1655, AnyNumberIf(WasKicked)) on a
// green body. The engine runs the bullets in printed order (CR 608.2c) whatever order they were clicked in, and that
// order is observable: kicked with all three on one creature, the
// counters land first, so the life bullet reads the bigger power and
// the fight deals it. X is read at resolution over layered power.
//
// The fight bullet is one statement with two clauses; each clause's
// pick is read by its slot within this occurrence's own group, and a
// fighter that has left (or gone illegal) means no damage at all (CR
// 701.14b). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5e28749-18ea-4a2b-b7d9-905cf2913d4e",
		Name:         "Inscription of Abundance",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			effects.Kicker("{2}{G}"),
		},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Put two +1/+1 counters on target creature.",
				effects.TargetCreature("target creature"),
				inscriptionOfAbundanceCounters),
			effects.ModeDoing("Target player gains X life, where X is the greatest power among creatures they control.",
				effects.TargetPlayer("target player"),
				inscriptionOfAbundanceLife),
			effects.ModeDoing("Target creature you control fights target creature you don't control.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				),
				effects.FightTheModesTargets),
		).AnyNumberIf(effects.WasKicked),
	})
}

// inscriptionOfAbundanceCounters is the first bullet.
func inscriptionOfAbundanceCounters(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 2}.Apply(ctx)
}

// inscriptionOfAbundanceLife is the second: X read now, over layered
// power, so counters the first bullet just placed count.
func inscriptionOfAbundanceLife(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	ctx.Game.RecomputeLayersIfStaleLocked()
	return effects.GainLife{Player: t.ID, Amount: effects.B42GreatestPowerControlledBy(ctx.Game, t.ID)}.Apply(ctx)
}
