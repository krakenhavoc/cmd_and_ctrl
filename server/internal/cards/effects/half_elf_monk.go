package effects

// Half-Elf Monk — Creature — Human Elf Monk {3}{W}, 1/4 (Adventures in
// the Forgotten Realms):
//
//	"Vigilance
//	 Stunning Strike — {1}{W}, {T}: Tap target creature."
//
// Arashin Sunshield's tapper. "Stunning Strike" has no rules meaning
// (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "00d11662-9d10-4915-b80c-c1ad54b4e71b",
		Name:            "Half-Elf Monk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "Stunning Strike — {1}{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
