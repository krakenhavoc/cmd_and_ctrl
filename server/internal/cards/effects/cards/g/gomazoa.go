package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gomazoa — Creature — Jellyfish {2}{U}, 0/3 (Commander 2011, #5574):
//
//	"Defender, flying
//	 {T}: Put this creature and each creature it's blocking on top of their
//	 owners' libraries, then those players shuffle their libraries."
//
// Guard Gomazoa's tuck without a target (Void Stalker's tail): as the ability
// resolves, Gomazoa and every attacker it blocks (CR 509.1a) that is still on
// the battlefield go on top of their owners' libraries together, and each
// owner shuffles once. Not blocking anything, it just tucks itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c82ae16-f995-4ac3-adb4-a5cc93a6e6f2",
		Name:            "Gomazoa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put this creature and each creature it's blocking on top of their owners' libraries, then those players shuffle their libraries.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Effect:  effects.S82OldB457TuckThisAndWhatItBlocks,
		}},
	})
}
