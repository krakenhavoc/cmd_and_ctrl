package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rolling Earthquake — Sorcery {X}{R} (Portal Three Kingdoms, S82 #5622):
//
//	"Rolling Earthquake deals X damage to each creature without horsemanship and
//	 each player."
//
// Earthquake's sweep with "without horsemanship" (OldB708DamageEachCreatureAndEachPlayer):
// the creatures are snapshotted before the first point lands, "without
// horsemanship" is read post-layer so a granted horsemanship keeps a creature out
// of it, and then every seated player, the caster included, takes X. Damage, not
// destruction: a creature that took lethal dies at the state-based sweep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "46b96552-830e-4732-9560-9872aeec858b",
		Name:     "Rolling Earthquake",
		// ADR 0126 §6: a creature with horsemanship is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB708DamageEachCreatureAndEachPlayer(effects.WithoutKeyword("horsemanship")),
	})
}
