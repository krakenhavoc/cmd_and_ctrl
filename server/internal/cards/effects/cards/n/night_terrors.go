package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Night Terrors — Sorcery {2}{B} (Innistrad, S82 older sets batch 449,
// #5572):
//
//	"Target player reveals their hand. You choose a nonland card from it.
//	 Exile that card."
//
// Check for Traps without the life clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d720de8-b7b7-478f-ab1a-3487ce76aa00",
		Name:         "Night Terrors",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.TargetRevealsYouChooseExile(effects.Nonland(), "nonland card"),
	})
}
