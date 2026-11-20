package effects

// Glass Golem — Artifact Creature — Golem {5}, 6/2 (Ravnica: City of Guilds,
// #5600):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24e48968-bfeb-4965-9e06-8b04397fcb53",
		Name:         "Glass Golem",
		Completeness: CompletenessFull,
	})
}
