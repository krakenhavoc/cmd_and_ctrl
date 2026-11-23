package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aeronaut Admiral — Creature — Human Pilot {3}{W}, 3/1 (S82 reprint pass):
//
//	"Flying
//	 Vehicles you control have flying."
//
// The grant is a layer-6 static over every Vehicle the controller has,
// crewed or not (Mu Yanling's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dcde1a62-33a5-4454-a389-9bf7a3521a4c",
		Name:            "Aeronaut Admiral",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.AstorVehicleYouControl, "flying"),
		},
	})
}
