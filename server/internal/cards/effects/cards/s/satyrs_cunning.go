package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Satyr's Cunning — Sorcery {R}:
//
//	"Create a 1/1 red Satyr creature token with "This token can't block."
//	 Escape—{2}{R}, Exile two other cards from your graveyard."
func init() {
	effects.Register(effects.Spec{
		OracleID:         "46fd9b9e-ef8f-41d6-9f1d-040bb7fe74c1",
		Name:             "Satyr's Cunning",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{2}{R}", 2)},
		OnResolve:        effects.ResolveSteps(effects.CreateToken{Template: effects.GapB104SatyrCantBlockToken(), N: 1}),
	})
}
