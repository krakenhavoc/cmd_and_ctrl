package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// green body. The bullets run in printed order (CR 608.2c), and that
// order is observable: kicked with all three on one creature, the
// counters land first, so the life bullet reads the bigger power and
// the fight deals it. X is read at resolution over layered power.
//
// The fight bullet is one statement with two clauses; each clause's
// pick is read by its slot within this occurrence's own group, and a
// fighter that has left (or gone illegal) means no damage at all (CR
// 701.14b). No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5e28749-18ea-4a2b-b7d9-905cf2913d4e",
		Name:         "Inscription of Abundance",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			Kicker("{2}{G}"),
		},
		Modes: ChooseOne(
			ModeDoing("Put two +1/+1 counters on target creature.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return AddCounter{Target: t.ID, Kind: "+1/+1", N: 2}.Apply(ctx)
				}),
			ModeDoing("Target player gains X life, where X is the greatest power among creatures they control.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return GainLife{Player: t.ID, Amount: b42GreatestPowerControlledBy(ctx.Game, t.ID)}.Apply(ctx)
				}),
			ModeDoing("Target creature you control fights target creature you don't control.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature you don't control", OpponentControls()),
				),
				inscriptionOfAbundanceFight),
		).AnyNumberIf(WasKicked),
	})
}

// inscriptionOfAbundanceFight is the third bullet: this occurrence's
// slot-0 pick fights its slot-1 pick, and only while both are still
// legal targets (CR 608.2b).
func inscriptionOfAbundanceFight(_ *game.StackItem, ctx *Context, occ int) error {
	var mine, theirs game.TargetRef
	var haveMine, haveTheirs bool
	for _, t := range ctx.ModeTargets(occ) {
		if !ctx.IsTargetLegal(t) || t.Kind != game.TargetCard {
			continue
		}
		switch t.Slot {
		case 0:
			mine, haveMine = t, true
		case 1:
			theirs, haveTheirs = t, true
		}
	}
	if !haveMine || !haveTheirs {
		return nil
	}
	return b10Fight(ctx, mine.ID, theirs.ID)
}
