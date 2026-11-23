package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weight of Conscience — Enchantment — Aura {1}{W} (Jumpstart 2022, #5948):
//
//	"Enchant creature
//	 Enchanted creature can't attack.
//	 Tap two untapped creatures you control that share a creature type: Exile
//	 enchanted creature."
//
// Forced Worship's restriction, and an activated ability paid by tapping two
// other creatures. That cost is not {T}, so summoning sickness does not stop
// a creature from paying it (CR 302.6). "Share a creature type"
// (TapOthersCost.SharesCreatureType) is judged by game.SharesCreatureType on
// the named pair, so a changeling shares with any creature that has a type,
// and the menu row is greyed when no two of your untapped creatures share
// one. The exile reads what the Aura enchants as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42989c6d-75d9-4ddd-bb9f-f18c0eb005fe",
		Name:         "Weight of Conscience",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttack)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two untapped creatures you control that share a creature type: Exile enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.GapB170TapTwoSharingATypeCost(),
			Effect:  effects.R41ExileEnchanted,
		}},
	})
}
