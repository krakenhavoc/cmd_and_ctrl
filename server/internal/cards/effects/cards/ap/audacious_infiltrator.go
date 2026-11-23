package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Audacious Infiltrator — Creature — Dwarf Rogue {1}{W}, 3/1 (Aether Revolt,
// #5536):
//
//	"This creature can't be blocked by artifact creatures."
//
// Argothian Sprite's evasion: a block rule on the creature itself that
// refuses an artifact creature as its blocker (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e01b3d52-297d-42e3-9de9-4f45c2359bcc",
		Name:         "Audacious Infiltrator",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.Artifact()), "artifact creatures"),
		},
	})
}
