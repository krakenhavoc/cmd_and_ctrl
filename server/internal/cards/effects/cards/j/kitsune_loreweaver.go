package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kitsune Loreweaver — Creature — Fox Cleric {1}{W}, 2/1 (Saviors of
// Kamigawa, #5601):
//
//	"{1}{W}: This creature gets +0/+X until end of turn, where X is the number
//	 of cards in your hand."
//
// X is the hand as the ability resolves (CR 608.2h), counted once: the
// boost does not follow the hand afterwards. The ability's own cost is paid
// first, so the mana is not a card in hand either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d989d8aa-2b2a-4fd2-a8c9-1d182909987c",
		Name:         "Kitsune Loreweaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: This creature gets +0/+X until end of turn, where X is the number of cards in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.OldB588ThisGetsPlusZeroPlusHandSize("Kitsune Loreweaver — +0/+X"),
		}},
	})
}
