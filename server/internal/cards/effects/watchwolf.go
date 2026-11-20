package effects

// Watchwolf — Creature — Wolf {G}{W}, 3/3 (Ravnica: City of Guilds, #5600):
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
		OracleID:     "a35c2e20-eb90-4132-b65f-be1fcb569819",
		Name:         "Watchwolf",
		Completeness: CompletenessFull,
	})
}
