package effects

// Your Fate Ends Here — Instant {2}{W} (Reality Fracture, tracker #2795):
//
//	"Destroy target creature or planeswalker with mana value 3 or greater.
//	 Surveil 1. (Look at the top card of your library. You may put it into
//	 your graveyard.)"
//
// A token has mana value 0 and so is never a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28a362c2-0c99-48e0-b9cb-485a6c1f350b",
		Name:         "Your Fate Ends Here",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target creature or planeswalker with mana value 3 or greater",
			And(Or(Creature(), Planeswalker()), ManaValueGE(3))),
		OnResolve: destroyTargetThenSurveil(1),
	})
}
