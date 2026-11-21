package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hinterland Drake — Creature — Drake {2}{U}, 2/3 (Aether Revolt,
// #5536):
//
//	"Flying
//	 This creature can't block artifact creatures."
//
// Cloud Djinn's blocker-side rule with the other attacker set: the pair
// is refused when the attacker is an artifact creature (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f457f182-69d0-468d-9bf0-7170363bac26",
		Name:            "Hinterland Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.Artifact())), "it can't block artifact creatures"),
		},
	})
}
