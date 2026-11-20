package effects

// Curious Pair // Treats to Share — Creature — Human Peasant {1}{G}, 1/3 // Sorcery — Adventure {G} (Throne of Eldraine, #5516):
//
//	Curious Pair — no rules text.
//	Treats to Share — "Create a Food token."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const curiousPairOracleID = "e67bd8eb-66db-48af-82f8-74f5595f1928"

func init() {
	Register(Spec{
		OracleID:     curiousPairOracleID,
		Name:         "Curious Pair",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     curiousPairOracleID + "#1",
		Name:         "Treats to Share",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(CreateToken{Template: FoodToken(), N: 1}),
	})
}
