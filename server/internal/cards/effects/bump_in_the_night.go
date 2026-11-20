package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bump in the Night — Sorcery {B} (Innistrad, #5572):
//
//	"Target opponent loses 3 life.
//	 Flashback {5}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Shadow Slice's life loss (shadowSliceLoseLife) with the shared
// Flashback: cast from the graveyard for {5}{R}, then exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d9f2c571-b357-4d4d-94c5-fd47551bb842",
		Name:             "Bump in the Night",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target opponent", Opponent()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{R}")},
		OnResolve:        shadowSliceLoseLife,
	})
}
