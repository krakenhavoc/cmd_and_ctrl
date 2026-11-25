package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maulfist Revolutionary — {1}{G}{G} Creature — Human Warrior 3/3 (Aether
// Revolt, #5536):
//
//	"Trample
//	 When this creature enters or dies, for each kind of counter on target
//	 permanent or player, give that permanent or player another counter of
//	 that kind."
//
// One printed ability, one registered trigger (it fires on either event).
// The shared body is Skyship Plunderer's: every kind of counter on the
// target gets one more, read as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65d41af0-978e-4071-9f64-0e3b5826763a",
		Name:            "Maulfist Revolutionary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrDies("Maulfist Revolutionary — for each kind of counter on target permanent or player, give it another counter of that kind",
					effects.MshA6AnotherOfEachKindOnTarget),
				effects.ArdennTargetPermanentOrPlayer()),
		},
	})
}
