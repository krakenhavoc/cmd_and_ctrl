package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apostle of Purifying Light — Creature — Human Cleric {1}{W}, 2/1 (Core Set 2020, #5518):
//
//	"Protection from black (This creature can't be blocked, targeted, dealt
//	 damage, enchanted, or equipped by anything black.)
//	 {2}: Exile target card from a graveyard."
//
// The target is any card in any graveyard; there is no {T}, so it can be
// activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b49dbdc-4f1d-4895-b263-a48c7b7e6b44",
		Name:            "Apostle of Purifying Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Exile target card from a graveyard.",
			Cost:    effects.ManaCost("{2}"),
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.ExileFirstTarget,
		}},
	})
}
