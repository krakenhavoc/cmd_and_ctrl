package effects

// Duskwielder — Creature — Elf Berserker {B}, 1/2:
//
//	"Boast — {1}: Target opponent loses 1 life and you gain 1 life. (Activate only if
//	 this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9227f38-7db7-4ea1-9482-94e5de3984fe",
		Name:         "Duskwielder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			BoastTargeting("{1}: Target opponent loses 1 life and you gain 1 life.",
				ManaCost("{1}"), TargetPlayer("target opponent", Opponent()), drainTargetOpponentOne),
		},
	})
}
