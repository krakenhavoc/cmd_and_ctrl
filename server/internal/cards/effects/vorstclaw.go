package effects

// Vorstclaw — {4}{G}{G} Creature — Elemental Horror 7/7 (Avacyn Restored, #5570):
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
		OracleID:     "aa944cbe-fdba-4001-97d5-c722fe744dcc",
		Name:         "Vorstclaw",
		Completeness: CompletenessFull,
	})
}
