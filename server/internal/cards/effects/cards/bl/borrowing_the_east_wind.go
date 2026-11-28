package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Borrowing the East Wind — Sorcery {X}{G}{G} (Portal Three Kingdoms, S82
// #5622):
//
//	"Borrowing the East Wind deals X damage to each creature with horsemanship and
//	 each player."
//
// Rolling Earthquake's sweep turned around (OldB708DamageEachCreatureAndEachPlayer):
// only creatures with horsemanship, read post-layer, and every seated player,
// the caster included, take X. Damage, not destruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "79334732-8db1-4393-9fed-3cde3c572c5c",
		Name:     "Borrowing the East Wind",
		// ADR 0126 §6: only creatures with horsemanship are hit.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB708DamageEachCreatureAndEachPlayer(effects.HasKeyword("horsemanship")),
	})
}
