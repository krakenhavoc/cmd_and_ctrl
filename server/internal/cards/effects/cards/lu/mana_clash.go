package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mana Clash — Sorcery {R} (Eighth Edition, S82 OLD batch 611, #5607):
//
//	"You and target opponent each flip a coin. Mana Clash deals 1 damage to
//	 each player whose coin comes up tails. Repeat this process until both
//	 players' coins come up heads on the same flip."
//
// Uncalled flips (CR 705.1: only the face matters), the caster's and then the
// opponent's, through the engine's coin rules (Krark's Thumb and the like).
// Each round's damage is one damage instance, and the rounds repeat until a
// round has both heads. With no target left the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2f5a5fd-14dc-48e8-bdb0-984a83288023",
		Name:         "Mana Clash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.OldB611ManaClash,
	})
}
