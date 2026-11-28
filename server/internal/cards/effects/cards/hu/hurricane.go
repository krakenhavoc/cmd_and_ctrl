package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hurricane — Sorcery {X}{G} (Portal Second Age, S82 older sets batch 736, #5625):
//
//	"Hurricane deals X damage to each creature with flying and each player."
//
// Rolling Earthquake's sweep with the creature filter "flying", read through the layers; the creatures are snapshotted before the first point lands and every player, the caster included, takes X as one damage instance (an X of 0 deals nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c021685-4017-49c7-9f58-2ae0243361a0",
		Name:         "Hurricane",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a creature without flying is spared.
		Purpose:   game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		XMatters:  true,
		OnResolve: effects.OldB708DamageEachCreatureAndEachPlayer(effects.HasKeyword("flying")),
	})
}
