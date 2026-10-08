package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const oteclanLandmarkOracleID = "2420828e-b886-42d3-81ba-8ee81b6e1eb8"

func init() {
	Register(Spec{
		OracleID:     oteclanLandmarkOracleID,
		Name:         "Oteclan Landmark",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Oteclan Landmark — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {2}{W}", "{2}{W}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:        oteclanLandmarkOracleID + "#1",
		Name:            "Oteclan Levitator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Oteclan Levitator — target attacking creature without flying gains flying",
					firstTargetGainsUntilEndOfTurn(0, "Oteclan Levitator — flying until end of turn", "flying")),
				TargetCreature("target attacking creature without flying", AttackingCreature(), WithoutKeyword("flying"))),
		},
	})
}
