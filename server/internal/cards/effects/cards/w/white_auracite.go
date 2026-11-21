package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// White Auracite — Artifact {2}{W}{W}:
//
//	"When this artifact enters, exile target nonland permanent an
//	 opponent controls until this artifact leaves the battlefield.
//	 {T}: Add {W}."
//
// Perilous Snare's entry trigger on a mana rock: the CR 610.3 "until"
// return (exileChosenTargetUntilThisLeaves, #1729), so the card comes
// back when the artifact leaves, and never if the artifact left before
// the trigger resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6022608a-6cf2-45bd-adec-63211710a5ed",
		Name:         "White Auracite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets:   effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls()),
			Key:       "White Auracite — exile target nonland permanent an opponent controls until this artifact leaves the battlefield",
			Effect:    effects.ExileChosenTargetUntilThisLeaves("White Auracite — the exiled card returns when White Auracite leaves the battlefield"),
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
