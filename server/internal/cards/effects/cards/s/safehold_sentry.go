package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Safehold Sentry — Creature — Elf Warrior {1}{W} 2/2:
//
//	"{2}{W}, {Q}: This creature gets +0/+2 until end of turn. ({Q} is the untap
//	 symbol.)"
//
// {Q} is the untap-self cost (Umbral Mantle's): the Sentry must be tapped to
// activate, is untapped as a cost, and summoning sickness does not apply to it
// (CR 302.6).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22534fab-2c62-4113-bcea-f305ae9f664f",
		Name:         "Safehold Sentry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {Q}: This creature gets +0/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 2}},
			Cost:    game.AbilityCost{Mana: "{2}{W}", UntapSelf: true},
			Effect:  effects.ThisGetsUntilEndOfTurn(0, 2, "Safehold Sentry — +0/+2 until end of turn"),
		}},
	})
}
