package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mischievous Chimera — Enchantment Creature — Chimera {U}{R}, 2/2 (Theros
// Beyond Death, #5514):
//
//	"Flying
//	 Whenever you cast your first spell during each opponent's turn, this creature
//	 deals 1 damage to each opponent. Scry 1."
//
// The same condition as Dreamstalker Manticore (B22FirstSpellOnAnOpponentsTurn).
// The damage is one event from the Chimera to each opponent, then the scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5cc57770-1070-45f6-9d76-d402c55a0d06",
		Name:            "Mischievous Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCastFirstSpellOnAnOpponentsTurn("Mischievous Chimera — 1 damage to each opponent, then scry 1",
				effects.Do(effects.DamageEachOpponent{N: 1}, effects.Scry{N: 1})),
		},
	})
}
