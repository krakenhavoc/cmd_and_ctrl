package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6022608a-6cf2-45bd-adec-63211710a5ed",
		Name:         "White Auracite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
			Key:       "White Auracite — exile target nonland permanent an opponent controls until this artifact leaves the battlefield",
			Effect:    exileChosenTargetUntilThisLeaves("White Auracite — the exiled card returns when White Auracite leaves the battlefield"),
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
