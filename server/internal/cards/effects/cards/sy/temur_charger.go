package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temur Charger — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Horse {1}{G}, 3/1: Morph—Reveal a green card in your hand. When this
//	creature is turned face up, target creature gains trample until end of turn.

func init() {
	effects.Register(effects.Spec{
		OracleID:         "a0665c10-e652-4608-920b-66206c0e5905",
		Name:             "Temur Charger",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.MorphReveal("G")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Temur Charger — target creature gains trample until end of turn",
					effects.EbTargetCreatureGainsUntilEOT("trample", "Temur Charger — trample until end of turn")),
				effects.TargetCreature("target creature")),
		},
	})
}
