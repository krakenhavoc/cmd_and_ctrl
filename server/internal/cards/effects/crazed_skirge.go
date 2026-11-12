package effects

// Crazed Skirge — Creature — Phyrexian Imp {3}{B}, 2/2 (Urza's Saga,
// #5624):
//
//	"Flying, haste"
//
// Registered so the creature resolves and enters on its own rather than
// through the manual sandbox path. Both keywords are printed card data:
// flying for the block check (CR 702.9a), haste so it can attack and tap
// the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72a53d9e-0645-4024-a90f-73b05c2284c6",
		Name:            "Crazed Skirge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
