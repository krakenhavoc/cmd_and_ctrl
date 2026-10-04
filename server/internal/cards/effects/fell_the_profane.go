package effects

// Fell the Profane // Fell Mire — modal double-faced card. This file
// is the FRONT face, Instant {2}{B}{B}:
//
//	"Destroy target creature or planeswalker."
//
// The back face, Fell Mire, is registered with the MDFC land cycle in
// mdfc_lands.go under "<oracle>#1".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "053a69d8-2b5e-4f14-8b02-ca405891dc4a",
		Name:         "Fell the Profane",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
