package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oteclan Landmark // Oteclan Levitator — a transforming artifact
// (#2124, ADR 0137):
//
//	Oteclan Landmark — Artifact {W}
//	  "When this artifact enters, scry 2.
//	   Craft with artifact {2}{W}"
//	Oteclan Levitator — Artifact Creature — Golem, 1/4
//	  "Flying
//	   Whenever this creature attacks, target attacking creature without
//	   flying gains flying until end of turn."
//
// Simulacrum Synthesizer's entry scry, and Eddytrail Hawk's grant
// without the energy. The Levitator itself has flying, so it is never
// its own target.
//
// No simplification.
const OteclanLandmarkOracleID = "2420828e-b886-42d3-81ba-8ee81b6e1eb8"

func init() {
	effects.Register(effects.Spec{
		OracleID:     OteclanLandmarkOracleID,
		Name:         "Oteclan Landmark",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oteclan Landmark — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with artifact {2}{W}", "{2}{W}", effects.CraftWith("artifact")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        OteclanLandmarkOracleID + "#1",
		Name:            "Oteclan Levitator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Oteclan Levitator — target attacking creature without flying gains flying",
					effects.FirstTargetGainsUntilEndOfTurn(0, "Oteclan Levitator — flying until end of turn", "flying")),
				effects.TargetCreature("target attacking creature without flying", effects.AttackingCreature(), effects.WithoutKeyword("flying"))),
		},
	})
}
