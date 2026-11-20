package effects

// Ancient Carp — Creature — Fish {4}{U}, 2/5 (Dragons of Tarkir,
// #5551):
//
//	"(no rules text)"
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84122d1e-7657-42fa-8724-c4a3a19729a6",
		Name:         "Ancient Carp",
		Completeness: CompletenessFull,
	})
}
