package effects

// Anurid Murkdiver — Creature — Zombie Frog Beast {4}{B}{B}, 4/3
// (Onslaught):
//
//	"Swampwalk"
//
// Swampwalk is a printed keyword, enforced by the block check
// (game/landwalk.go): it can't be blocked while the defending player
// controls a Swamp.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7dc3d8b-2c79-463d-b862-9f01ccea05fd",
		Name:            "Anurid Murkdiver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
