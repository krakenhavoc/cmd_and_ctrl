package effects

// Quarrel — Instant {1}{G} (The Hobbit, HOB):
//
//	"Target creature you control deals damage equal to its power to
//	 target creature an opponent controls."
//
// Bite Down's shape at the narrower recipient: two target clauses read
// positionally (#764), slot 0 the damage source (its power read as the
// spell resolves, its deathtouch and lifelink carried by the damage) and
// slot 1 the recipient, each re-checked on its own at resolution
// (CR 608.2b). The creature deals the damage, so nothing comes back
// (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db05b114-2eb5-46bc-81e7-7e036d8eaab3",
		Name:         "Quarrel",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: oneSidedBite,
	})
}
