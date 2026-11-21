package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Bauble — Artifact {B} (Aetherdrift, #4464):
//
//	"When this artifact enters, target creature an opponent controls gets
//	 -2/-2 until end of turn.
//	 {2}{B}, {T}, Sacrifice this artifact: Surveil 2. (Look at the top two
//	 cards of your library, then put any number of them into your
//	 graveyard and the rest on top of your library in any order.)"
//
// Ambush Gigapede's enters shrink on an artifact, then Broodheart Engine's
// sacrifice-and-tap cost paying for a plain surveil 2 (CR 701.25a). The
// shrink's target is chosen as the trigger goes on the stack and
// re-checked as it resolves (CR 608.2b); with no opponent's creature the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "431d6a2b-e22c-4921-a5f1-dd672f764f78",
		Name:         "Grim Bauble",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Grim Bauble — target creature an opponent controls gets -2/-2 until end of turn",
				effects.BoostTargetUntilEOT("Grim Bauble — -2/-2", -2, -2)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}, Sacrifice this artifact: Surveil 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.Surveil{N: 2}),
		}},
	})
}
