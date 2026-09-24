package effects

// Puncture Blast — Instant {2}{R}:
//
//	"Wither
//	 Puncture Blast deals 3 damage to any target."
//
// The card that proves wither works from the stack (CR 702.80c, ADR
// 0056 Decision 2): the damage tail reads a spell's printed keywords
// off the spell while it resolves, so the 3 damage becomes three
// -1/-1 counters on a creature. Wither says nothing about players, so
// at a player it is 3 life lost. Damage to a planeswalker takes
// loyalty as usual.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "e296581d-01ac-43bf-898c-2edb4c81bcbe",
		Name:            "Puncture Blast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"wither"},
		Targets:         TargetAny(),
		OnResolve:       damageToFirstTarget(3),
	})
}
