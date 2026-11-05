package effects

// Fleet Captain Christopher Pike — Legendary Creature — Human Officer
// {2}{U}, 0/6 (Star Trek Commander, #4974):
//
//	"Yes — {1}, {T}: Untap two other target creatures.
//	 No — {2}, {T}: Tap two other target creatures."
//
// "Yes" and "No" are ability words (CR 207.2c). Each ability needs exactly
// two targets, neither of them Pike ("other": Another). A target that
// became illegal is skipped and the other is still tapped or untapped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0a9a7a2-c996-4e3e-817d-9283be690c39",
		Name:         "Fleet Captain Christopher Pike",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Yes — {1}, {T}: Untap two other target creatures.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: Another(TargetCreature("two other target creatures").WithCount(2, 2)),
				Effect:  UntapEachLegalTarget,
			},
			{
				Label:   "No — {2}, {T}: Tap two other target creatures.",
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Targets: Another(TargetCreature("two other target creatures").WithCount(2, 2)),
				Effect:  tapEachLegalTarget,
			},
		},
	})
}
