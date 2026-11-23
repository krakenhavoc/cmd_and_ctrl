package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream-Thief's Bandana — Artifact — Equipment {2} (Outlaws of Thunder
// Junction Commander):
//
//	"Whenever equipped creature deals combat damage to a player, look at the
//	 top card of their library, then exile it face down. For as long as it
//	 remains exiled, you may play it, and mana of any type can be spent to
//	 cast that spell.
//	 Equip {1}"
//
// Gonti, Night Minister's face-down exile with the any-type grant, to the
// Bandana's controller, off the library of the player the equipped creature
// hit. "Play", so a land exiled this way can be played as a land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c958b5f-680b-45fc-8c3d-27b23677436e",
		Name:         "Dream-Thief's Bandana",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.Cmdr2S45EquippedCreatureDealtCombatDamageToAPlayer,
				"Dream-Thief's Bandana — look at the top card of their library, then exile it face down; you may play it",
				effects.Cmdr2S45ExileTopFaceDownToPlay),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
