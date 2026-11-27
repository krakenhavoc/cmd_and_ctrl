package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Distortion — Creature — Wall {2}{B}{B}, 1/3 (Mercadian Masques, S82
// OLD b682, #5619):
//
//	"Defender
//	 {2}{B}, {T}: Target player discards a card. Activate only as a sorcery."
//
// Defender is a printed keyword. Cinderhaze Wretch's discard on a target
// player (who chooses the card, CR 701.9a), gated to the sorcery-speed window
// at activation (CR 602.5d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cc9cbbbc-8b02-42ef-a4d8-e2837e742c88",
		Name:            "Wall of Distortion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}, {T}: Target player discards a card. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target player"),
			Effect:       effects.TargetPlayerDiscards(1),
		}},
	})
}
