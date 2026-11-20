package effects

// Noxious Grasp — Instant {1}{B} (Core Set 2020, #5518):
//
//	"Destroy target creature or planeswalker that's green or white. You gain 1
//	 life."
//
// Terminal Criticism's body over green or white.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b17f4960-9603-4a81-8dc8-e09e8b822357",
		Name:         "Noxious Grasp",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target creature or planeswalker that's green or white",
			And(Or(Creature(), Planeswalker()), Or(OfColor("G"), OfColor("W")))),
		OnResolve: socB26DestroyTheTargetAndGainLife(1),
	})
}
