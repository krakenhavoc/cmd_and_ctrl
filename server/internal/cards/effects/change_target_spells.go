package effects

// The three instants whose whole text is "Change the target of target
// spell with a single target." — Deflection {3}{U}, Swerve {U}{R},
// Shunt {1}{R}{R}. A table rather than three files because the cards
// are word-for-word the same ability; only the name and id differ, and
// the price lives on the card, not here.
//
// "With a single target" is the clause predicate HasASingleTarget, so a
// two-target spell is never offered to the picker (CR 115.7b). They
// print "target SPELL", not "target spell or ability", so there is no
// ability half to reach. The new target is chosen by the controller
// over the same legality gate the original announce ran (hexproof,
// protection, the clause's own restriction), and when no other legal
// target exists the target does not change. Nothing is simplified.
func init() {
	for _, c := range []struct{ id, name string }{
		{"ec7ae9ed-dc5b-47ed-a4ad-086f3c7c377c", "Deflection"},
		{"c68627b1-025c-48c9-9646-98eb2d268e71", "Swerve"},
		{"e714bcfb-b451-4b7b-ab60-4cb845c75647", "Shunt"},
	} {
		Register(Spec{
			OracleID:     c.id,
			Name:         c.name,
			Completeness: CompletenessFull,
			Targets:      TargetSpell("target spell with a single target", HasASingleTarget()),
			OnResolve:    changeTheTarget(c.name + " — change the target"),
		})
	}
}
