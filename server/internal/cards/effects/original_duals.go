package effects

// original_duals.go — the four most-played of the Alpha "dual lands":
//
//	"({T}: Add {U} or {B}.)"
//
// Land — Island Swamp, with the mana ability as reminder text
// because the basic land TYPES carry it. That is exactly why they
// need a catalog entry: the engine synthesises a mana ability from
// the type line only for the BASIC supertype (ManaAbilitiesForCard →
// basicLandColor), so an unregistered Underground Sea produces
// nothing at all — the same trap the battle lands fell into. One
// pipe ability, no enters-tapped, no drawback: the best lands in the
// format are also the simplest.
//
// Roadmap batch 02 (#295) ranks these four; the other six duals sit
// in later batches and belong in this table when they arrive.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"4b22be3a-8ce1-47d1-b82e-6c3ccfb0548b", "Underground Sea", "U", "B"},
		{"c718911c-c955-4eb9-9e16-be4bd49a4e4e", "Volcanic Island", "U", "R"},
		{"74b7fe23-5d3a-4092-8d78-7c0eba8f6f73", "Tropical Island", "G", "U"},
		{"02418479-9455-417f-a6a1-004356faff37", "Tundra", "W", "U"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
