package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saprazzan Outrigger — Creature — Merfolk {3}{U}, 5/5 (Mercadian Masques, S82 OLD batch 688, #5619):
//
//	"When this creature attacks or blocks, put it on top of its owner's library at end of combat."
//
// Ceremonial Guard's attack-or-block trigger (one trigger however many attackers it blocks) over a
// delayed trigger at the beginning of the end of combat step (CR 511.2) that tucks the Outrigger,
// skipped if it has left or come back as a new object (CR 400.7). A keyed body, on disk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd57638e-8444-4522-8638-4514dec8b3ed",
		Name:         "Saprazzan Outrigger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
				"Saprazzan Outrigger — put it on top of its owner's library at end of combat",
				effects.OldB688TuckThisAtEndOfCombat("Saprazzan Outrigger — put it on top of its owner's library")),
		},
	})
}
