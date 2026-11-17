package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scattershot — Instant {2}{R} (Scourge):
//
//	"Scattershot deals 1 damage to target creature.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Grapeshot with a creature-only target clause. Each copy is its own
// one-damage instruction and may point at a different creature (CR
// 702.40a, CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed3652da-0127-4c9e-9913-f20becdd0ba9",
		Name:         "Scattershot",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 1)),
		Targets:      TargetCreature("target creature"),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve:    damageToFirstTarget(1),
	})
}
