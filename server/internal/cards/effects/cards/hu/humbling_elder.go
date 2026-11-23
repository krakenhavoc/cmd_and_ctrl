package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Humbling Elder — Creature — Human Monk {U}, 1/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"Flash
//	 When this creature enters, target creature an opponent controls gets
//	 -2/-0 until end of turn."
//
// Cogwork Wrestler's text on a Monk. The trigger is removed with no legal
// target (CR 603.3d), and the shrink ends in the cleanup step (CR 514.2).
// A target no longer legal is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "013bcab8-e527-4700-82c9-4cb9cb8a4807",
		Name:            "Humbling Elder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Humbling Elder — target creature an opponent controls gets -2/-0 until end of turn",
				effects.BoostTargetUntilEOT("Humbling Elder — -2/-0 until end of turn", -2, 0)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, 0))),
		},
	})
}
