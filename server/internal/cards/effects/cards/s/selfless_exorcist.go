package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selfless Exorcist — Creature — Human Cleric {3}{W}{W}, 3/4 (Judgment,
// #5611):
//
//	"{T}: Exile target creature card from a graveyard. That card deals
//	 damage equal to its power to this creature."
//
// The power is read from the card where it lay, before it is exiled (CR
// 608.2h), and the exiled card is the damage's source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78e875b1-e621-4bf0-ad81-24ed9e2b43f8",
		Name:         "Selfless Exorcist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Exile target creature card from a graveyard. That card deals damage equal to its power to this creature.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Effect:  effects.OldB621ExileTargetCreatureCardItDealsItsPower,
		}},
	})
}
