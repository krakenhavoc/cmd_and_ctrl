package effects

// Mothrider Patrol — Creature — Fox Warrior {W}, 1/1 (Neon Dynasty):
//
//	"Flying
//	 {3}{W}, {T}: Tap target creature."
//
// Flying is the printed keyword; the tapper is Arashin Sunshield's body
// with a larger mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c4073754-7d92-4976-8e7d-d6eb85b91608",
		Name:            "Mothrider Patrol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{3}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
