package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Excelsior, Cutting Edge — Legendary Artifact — Spacecraft {3}{W},
// 4/4 (Star Trek):
//
//	When U.S.S. Excelsior enters, destroy target tapped creature an
//	opponent controls.
//	Station (Tap another creature you control: Put charge counters
//	equal to its power on this Spacecraft. Station only as a sorcery.
//	It's an artifact creature at 8+.)
//	8+ | Flying
//
// The enters trigger targets Henchbots' way ("target tapped creature an
// opponent controls"); a target that untapped or changed control before
// resolution is illegal and the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f4ae54b-f77d-43df-87d5-792ac0905058",
		Name:         "U.S.S. Excelsior, Cutting Edge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("U.S.S. Excelsior — destroy target tapped creature an opponent controls", destroyFirstLegalTarget),
				TargetCreature("target tapped creature an opponent controls", tappedPermanent(), OpponentControls())),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 4, 4),
			ThresholdKeywords(8, "flying"),
		),
	})
}
