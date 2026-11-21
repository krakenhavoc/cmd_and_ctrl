package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winter Soldier, Bucky Barnes — Legendary Creature — Human Soldier Hero
// {W}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Winter Soldier enters tapped."
//
// SelfEntersTapped (CR 614.1c): a replacement that applies as he enters,
// so he never untaps for a moment and no "becomes tapped" trigger sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "874bbeb2-aa15-41e8-afad-122ab9b555fd",
		Name:         "Winter Soldier, Bucky Barnes",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
