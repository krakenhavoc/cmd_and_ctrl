package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tilonalli's Skinshifter — Creature — Human Shaman {2}{R}, 0/1 (Ixalan, S82
// older sets):
//
//	"Haste
//	 Whenever this creature attacks, it becomes a copy of another target
//	 nonlegendary attacking creature until end of turn."
//
// Impossible Man's copy (SelfBecomesCopyOfTarget, until end of turn): the
// target is chosen as the trigger goes on the stack (CR 603.3d) from the
// other nonlegendary attackers, and is checked again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23df9146-9fce-4769-b017-87aa36c79300",
		Name:            "Tilonalli's Skinshifter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Tilonalli's Skinshifter — it becomes a copy of another target nonlegendary attacking creature until end of turn",
				effects.SelfBecomesCopyOfTarget("Tilonalli's Skinshifter — becomes a copy until end of turn")),
				effects.Another(effects.TargetCreature("another target nonlegendary attacking creature",
					effects.AttackingCreature(), effects.Not(effects.Legendary())))),
		},
	})
}
