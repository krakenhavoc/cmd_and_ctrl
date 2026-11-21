package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stingblade Assassin — Creature — Faerie Assassin {3}{B}, 3/1 (Wilds of
// Eldraine, #4471):
//
//	"Flash
//	 Flying
//	 When this creature enters, destroy target creature an opponent
//	 controls that was dealt damage this turn."
//
// Flash and flying are the engine's keywords. The entry trigger is Covert
// Cutpurse's: the shared DealtDamageThisTurn predicate on a creature an
// opponent controls, destroyed on resolution. With no such creature the
// trigger has no target and is removed from the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7938112a-3832-42b4-b123-27a42c06a290",
		Name:            "Stingblade Assassin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Stingblade Assassin — destroy target creature an opponent controls that was dealt damage this turn", effects.DestroyChosenPermanent),
				effects.TargetCreature("target creature an opponent controls that was dealt damage this turn", effects.OpponentControls(), effects.DealtDamageThisTurn())),
		},
	})
}
