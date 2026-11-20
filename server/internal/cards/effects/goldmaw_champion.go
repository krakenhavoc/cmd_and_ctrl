package effects

// Goldmaw Champion — Creature — Dwarf Warrior {2}{W}, 2/3 (Kaldheim,
// #5505):
//
//	"Boast — {1}{W}: Tap target creature. (Activate only if this
//	 creature attacked this turn and only once each turn.)"
//
// The boast constructor (boast.go, CR 702.142) carries the "attacked
// this turn, once each turn" gate; the body is the shared "Tap target
// creature" (tapChosenPermanent, Clockwork Drawbridge's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cbc8765-1735-48ca-bbc2-411aa7da2ed3",
		Name:         "Goldmaw Champion",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			BoastTargeting("{1}{W}: Tap target creature.",
				ManaCost("{1}{W}"), TargetCreature("target creature"), tapChosenPermanent),
		},
	})
}
