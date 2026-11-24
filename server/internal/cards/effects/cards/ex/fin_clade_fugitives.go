package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fin-Clade Fugitives — {5}{G} Creature — Elf Salamander Rogue 7/4, Commander
// Legends (#5507):
//
//	"This creature can't be blocked by creatures with power 2 or less.
//	 Encore {4}{G}"
//
// Rust-Shield Rampager's block rule, each blocker's power read as blockers are
// declared, and the shared Encore constructor (encore.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dad2965a-ceb4-42ac-bb62-de7eea500011",
		Name:         "Fin-Clade Fugitives",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Encore("{4}{G}")},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
