package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d1d719e9-9c02-40cf-8849-91a5c120a516",
		Name:         "Alabaster Host Intercessor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Targets:   effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Key:       "Alabaster Host Intercessor — exile target creature an opponent controls until this creature leaves the battlefield",
			Effect: effects.ExileChosenTargetUntilThisLeaves(
				"Alabaster Host Intercessor — the exiled creature returns when Alabaster Host Intercessor leaves the battlefield"),
		}},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}
