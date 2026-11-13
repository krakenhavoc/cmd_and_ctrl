package effects

// Cephalid Retainer — Creature — Octopus {2}{U}{U}, 2/3 (Odyssey, #5613):
//
//	"{U}{U}: Tap target creature without flying."
//
// Arashin Sunshield's tapper with no tap symbol, so it can be used again
// and again, and on the turn the Retainer arrives. A creature that gains
// flying in response is no longer a legal target (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a0a926c-59ab-4bd3-a238-13988bdd96a5",
		Name:         "Cephalid Retainer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}: Tap target creature without flying.",
			Cost:    ManaCost("{U}{U}"),
			Targets: TargetCreature("target creature without flying", Not(HasKeyword("flying"))),
			Effect:  tapChosenPermanent,
		}},
	})
}
