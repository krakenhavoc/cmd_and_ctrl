package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ice Flan — Creature — Elemental Ooze {4}{U}{U}, 5/4 (Final Fantasy,
// #4462):
//
//	"When this creature enters, tap target artifact or creature an
//	 opponent controls. Put a stun counter on it.
//	 Islandcycling {2}"
//
// Summon: Shiva's tap-and-stun over an artifact-or-creature target an
// opponent controls. Islandcycling is Angel of the Ruins's Plainscycling
// with the Island subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b7120a6-700b-4c61-bea4-a835f0cc0bcb",
		Name:         "Ice Flan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ice Flan — tap target artifact or creature an opponent controls and put a stun counter on it",
				effects.FinB1TapAndStun), effects.TargetPermanent("target artifact or creature an opponent controls", effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Islandcycling", "{2}", "an Island card", effects.IsLandWithSubtype("Island")),
		},
	})
}
