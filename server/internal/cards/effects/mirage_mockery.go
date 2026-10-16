package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirage Mockery — Sorcery {2}{U}:
//
//	"Choose one —
//	 • Create a token that's a copy of target artifact creature you
//	   control.
//	 • Create a token that's a copy of target nonartifact creature you
//	   control.
//	 Entwine {2}{U} (Choose both if you pay the entwine cost.)"
//
// Cackling Counterpart split along the artifact line, each bullet with
// its own target (TokenCopyTheModesTarget, CR 707.2's copiable values).
// Entwined (CR 702.42, #3866) it copies one of each, the artifact
// creature first (CR 702.42b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "2b0c08fb-7c46-484a-a446-b26a94db7172",
		Name:          "Mirage Mockery",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}{U}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Create a token that's a copy of target artifact creature you control.",
				TargetCreature("target artifact creature you control", YouControl(), Artifact()),
				TokenCopyTheModesTarget),
			ModeDoing("Create a token that's a copy of target nonartifact creature you control.",
				TargetCreature("target nonartifact creature you control", YouControl(), Not(Artifact())),
				TokenCopyTheModesTarget),
		)),
	})
}
