package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9f4ae54b-f77d-43df-87d5-792ac0905058",
		Name:         "U.S.S. Excelsior, Cutting Edge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("U.S.S. Excelsior — destroy target tapped creature an opponent controls", effects.DestroyFirstLegalTarget),
				effects.TargetCreature("target tapped creature an opponent controls", effects.TappedPermanent(), effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(8, 4, 4),
			effects.ThresholdKeywords(8, "flying"),
		),
	})
}
