package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storage Matrix — Artifact {3} (Urza's Destiny, S82 #5621):
//
//	"As long as this artifact is untapped, each player chooses artifact,
//	 creature, or land during their untap step. That player can untap only
//	 permanents of the chosen type this step."
//
// The engine's untap type choice (#6218, game.UntapTypeChoice): every player is
// asked at their own untap step, before anything untaps, while the Matrix is
// untapped as that step begins (read once, as Winter Orb's is); only permanents
// of the chosen type untap, and with two Matrices only a permanent of both chosen
// types does. A tapped Matrix asks nothing and untaps with everything else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03136fe1-17e2-4b1c-860a-dc6373c8fdb5",
		Name:         "Storage Matrix",
		Completeness: effects.CompletenessFull,
		UntapTypeChoices: []game.UntapTypeChoice{{
			Types:         []string{"artifact", "creature", "land"},
			WhileUntapped: true,
			Label:         "Storage Matrix",
		}},
	})
}
