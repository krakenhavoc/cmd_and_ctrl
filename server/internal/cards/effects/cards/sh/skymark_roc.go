package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skymark Roc — Creature — Bird {2}{W}{U}, 3/3 (Return to Ravnica, #5566):
//
//	"Flying
//	 Whenever this creature attacks, you may return target creature defending
//	 player controls with toughness 2 or less to its owner's hand."
//
// "Defending player" is the player the Roc is attacking (CR 802.2a); the toughness
// bound is read as the target is chosen and again on resolution (CR 608.2b). The
// "may" is asked on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7564801-e35c-4d5c-afa9-99b33b35309c",
		Name:            "Skymark Roc",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WithTargetsFrom(
				effects.WheneverThisAttacks("Skymark Roc — you may return target creature defending player controls with toughness 2 or less to its owner's hand", effects.BounceTheTargetToHand),
				effects.OldB424TargetSmallCreatureDefendingPlayerControls),
				"Skymark Roc — return target creature to its owner's hand?"),
		},
	})
}
