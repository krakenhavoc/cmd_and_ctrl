package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunhome Enforcer — Creature — Giant Soldier {2}{R}{W}, 2/4 (Ravnica: City of
// Guilds):
//
//	"Whenever this creature deals combat damage, you gain that much life.
//	 {1}{R}: This creature gets +1/+0 until end of turn."
//
// The trigger fires for each combat damage event this creature is the source of
// (to a player, a planeswalker or a creature), and gains its controller the
// amount that event carried, read as the ability resolves. The pump is
// firebreathing (Bellows Lizard's body): no {T}, so any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4c6f546-ad58-41ba-aadd-724958cd4611",
		Name:         "Sunhome Enforcer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ThisDealsCombatDamageYouGainThatMuchLife("Sunhome Enforcer")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Sunhome Enforcer — +1/+0 until end of turn"),
		}},
	})
}
