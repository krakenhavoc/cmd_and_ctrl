package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alabaster Host Intercessor — Creature — Phyrexian Samurai {5}{W}, 3/4
// (March of the Machine):
//
//	"When this creature enters, exile target creature an opponent controls
//	 until this creature leaves the battlefield.
//	 Plainscycling {2}"
//
// Brutal Cathar's enters trigger: the exile is keyed to this object (CR 610.3),
// so the creature comes back when this one leaves for any reason, and if it has
// already left by resolution nothing is exiled. Plainscycling is the shared
// Typecycling constructor (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1d719e9-9c02-40cf-8849-91a5c120a516",
		Name:         "Alabaster Host Intercessor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Targets:   TargetCreature("target creature an opponent controls", OpponentControls()),
			Key:       "Alabaster Host Intercessor — exile target creature an opponent controls until this creature leaves the battlefield",
			Effect: exileChosenTargetUntilThisLeaves(
				"Alabaster Host Intercessor — the exiled creature returns when Alabaster Host Intercessor leaves the battlefield"),
		}},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
