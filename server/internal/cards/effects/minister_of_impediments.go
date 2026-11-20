package effects

// Minister of Impediments — Creature — Human Advisor {2}{W/U}, 1/1 (Dissension, #5598):
//
//	"({W/U} can be paid with either {W} or {U}.)
//	 {T}: Tap target creature."
//
// Clockwork Drawbridge's tap row on a creature, so summoning sickness
// holds it back (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a51a0ee-bc81-4c42-bb17-a41cea957d9a",
		Name:         "Minister of Impediments",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
