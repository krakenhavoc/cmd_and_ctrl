package effects

// Drafna, Founder of Lat-Nam — Legendary Creature — Human Artificer
// Advisor {1}{U}, 2/1 (The Brothers' War promo):
//
//	"{1}{U}: Return target artifact you control to its owner's hand.
//	 {3}, {T}: Copy target artifact spell you control. (The copy becomes a
//	 token.)"
//
// The first is a plain bounce of one of the controller's artifacts. The
// second targets an artifact spell its controller controls on the stack
// (so the copy is made only while that spell is still there, CR 608.2b);
// the copy is controlled by Drafna's controller (CR 707.10), and a copy of
// a permanent spell becomes a token as it resolves (CR 707.10b). The
// copy's targets, if it has any (an artifact Aura), may be changed
// (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed2e8718-4b20-4310-96ce-05a29a24ef32",
		Name:         "Drafna, Founder of Lat-Nam",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{U}: Return target artifact you control to its owner's hand.",
				Cost:    ManaCost("{1}{U}"),
				Targets: TargetPermanent("target artifact you control", Artifact(), YouControl()),
				Effect:  s82Bro19BounceTarget,
			},
			{
				Label:   "{3}, {T}: Copy target artifact spell you control.",
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetSpell("target artifact spell you control", Artifact(), YouControl()),
				Effect:  copyTargetedSpell,
			},
		},
	})
}
