package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rally the Peasants — Instant {2}{W} (Innistrad, #5572):
//
//	"Creatures you control get +2/+0 until end of turn.
//	 Flashback {2}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Righteous Charge's team pump (creaturesYouControlGetUntilEOT) at
// +2/+0 with the shared Flashback ({2}{R}): the creatures you control as
// it resolves are the ones pumped (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3f0a0a15-613a-4864-9309-a340dc4cf94d",
		Name:             "Rally the Peasants",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{R}")},
		OnResolve:        effects.CreaturesYouControlGetUntilEOT(2, 0, "Rally the Peasants — creatures you control get +2/+0"),
	})
}
