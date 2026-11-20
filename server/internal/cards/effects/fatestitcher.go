package effects

// Fatestitcher — Creature — Zombie Wizard {3}{U}, 1/2 (Shards of Alara,
// #5588):
//
//	"{T}: You may tap or untap another target permanent.
//	 Unearth {U}"
//
// Trickster Mage's "you may tap or untap" body on Kiora's Follower's
// "another target permanent" clause: the controller picks tap it, untap
// it or leave it as the ability resolves, and a target gone in response
// is illegal, so nothing is asked (CR 608.2b). "Another" is object
// identity (Another), so the Fatestitcher can't pick itself. Unearth is
// the shared constructor (CR 702.84a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55765934-07be-4f03-a2e8-8292c0b601ae",
		Name:         "Fatestitcher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: You may tap or untap another target permanent.",
				Cost:    TapCost(),
				Targets: Another(TargetPermanent("another target permanent")),
				Effect:  nemMayTapOrUntapTheTarget("Fatestitcher"),
			},
			Unearth("{U}"),
		},
	})
}
