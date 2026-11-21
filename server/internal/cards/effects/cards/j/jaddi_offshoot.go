package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaddi Offshoot — Creature — Plant {G}, 0/3 (Tarkir: Dragonstorm
// Commander):
//
//	"Defender
//	 Landfall — Whenever a land you control enters, you gain 1 life."
//
// Eumidian Terrabotanist's landfall row on a defender: one trigger for
// each land that enters under your control, however it got there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20faefbd-059c-4aae-81e3-31683bf9f7bf",
		Name:            "Jaddi Offshoot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Jaddi Offshoot — you gain 1 life (landfall)", game.LandfallPayoff{LifeGain: 1}, effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
