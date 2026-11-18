package effects

// Skaab Wrangler — Creature — Human Wizard {1}{U}, 2/1 (Midnight Hunt, #5497):
//
//	"Tap three untapped creatures you control: Tap target creature."
//
// The cost is a TapOthers component, not the {T} symbol (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8de2ce5a-9345-4d2b-a248-868268f36836",
		Name:         "Skaab Wrangler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap three untapped creatures you control: Tap target creature.",
			Cost:    s82MidB09TapThreeCreatures(),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
