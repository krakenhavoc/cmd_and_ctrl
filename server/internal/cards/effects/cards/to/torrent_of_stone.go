package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torrent of Stone — Instant — Arcane {3}{R}:
//
//	"Torrent of Stone deals 4 damage to target creature.
//	 Splice onto Arcane—Sacrifice two Mountains. (As you cast an Arcane
//	 spell, you may reveal this card from your hand and pay its splice cost.
//	 If you do, add this card's effects to that spell.)"
//
// The splice cost sacrifices two Mountains, named as the spell is cast and
// sacrificed with it on the stack (#4055, ADR 0202's 2026-10-19 amendment),
// so a dies trigger lands above the spell. A land that is a Mountain by
// type, basic or not, pays; one that was only turned into a Mountain counts
// while it is one. Spliced onto another Arcane spell the damage is that
// spell's (CR 702.47c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc6ee33a-4245-4edb-a6b3-0bdcc2883fbc",
		Name:         "Torrent of Stone",
		Completeness: effects.CompletenessFull,
		Splice: effects.SpliceOntoArcaneCost(
			effects.SpliceSacrifice(2, "two Mountains", effects.Land(), effects.Subtype("Mountain"))),
		Targets: effects.TargetCreature("target creature"),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
		},
	})
}
