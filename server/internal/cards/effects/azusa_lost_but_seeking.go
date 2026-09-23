package effects

// Azusa, Lost but Seeking — Legendary Creature — Human Monk {2}{G},
// 1/2 (EDHREC rank 326):
//
//	"You may play two additional lands on each of your turns."
//
// Exploration's two-drop sibling — see exploration.go — and the
// other card `Spec.AdditionalLandPlays` was built for by name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "6c2c8bf3-9bf8-4a86-89d3-3bb36260dc51",
		Name:                "Azusa, Lost but Seeking",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 2,
	})
}
