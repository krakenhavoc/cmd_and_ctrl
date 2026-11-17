package effects

// Homarid Shaman — Creature — Homarid Shaman {2}{U}{U}, 2/1 (Fallen
// Empires, #5637):
//
//	"{U}: Tap target green creature."
//
// Ali Baba's tap with a colour in place of a subtype: the target must
// be green as the ability is activated and again as it resolves (CR
// 608.2b). No {T} in the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "137e3974-74e1-4941-8597-10aeba6cac37",
		Name:         "Homarid Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}: Tap target green creature.",
			Cost:    ManaCost("{U}"),
			Targets: TargetCreature("target green creature", OfColor("G")),
			Effect:  tapTheFirstLegalTarget,
		}},
	})
}
