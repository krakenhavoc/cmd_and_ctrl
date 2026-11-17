package effects

// Auriok Windwalker — Creature — Human Wizard {3}{W}, 2/3 (Fifth Dawn,
// #5604):
//
//	"Flying
//	 {T}: Attach target Equipment you control to target creature you
//	 control."
//
// Brass Squire's ability on a flier: two target clauses read back with
// AttachClauseTargets, at instant speed, since nothing on the card
// limits it to sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8a06210-de30-4e87-835e-80a6df5716ca",
		Name:            "Auriok Windwalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{T}: Attach target Equipment you control to target creature you control.",
			Cost:  TapCost(),
			Targets: TwoSlotAttachTargets(
				"target Equipment you control", []CardPredicate{YouControl()},
				"target creature you control", []CardPredicate{YouControl()},
			),
			Effect: AttachClauseTargets,
		}},
	})
}
