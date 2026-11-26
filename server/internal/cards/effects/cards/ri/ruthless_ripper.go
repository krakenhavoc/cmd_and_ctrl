package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruthless Ripper — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Human Assassin {B}, 1/1: Deathtouch. Morph—Reveal a black card in
//	your hand. When this creature is turned face up, target player loses 2 life.

func init() {
	effects.Register(effects.Spec{
		OracleID:         "cfbcd70b-d3a9-47e9-a8b3-b8134d00dff8",
		Name:             "Ruthless Ripper",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{effects.MorphReveal("B")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Ruthless Ripper — target player loses 2 life", effects.Do(effects.TargetPlayerLosesLife{N: 2})),
				effects.TargetPlayer("target player")),
		},
	})
}
