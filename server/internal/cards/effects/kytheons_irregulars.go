package effects

// Kytheon's Irregulars — Creature — Human Soldier {2}{W}{W}, 4/3:
//
//	"Renown 1 (When this creature deals combat damage to a player, if
//	 it isn't renowned, put a +1/+1 counter on it and it becomes
//	 renowned.)
//	 {W}{W}: Tap target creature."
//
// #2049: renown is the engine's keyword trigger (game/renown.go); the
// tapper is Arashin Sunshield's body with no {T} in its cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccfddb1a-1b7f-44f9-89c4-1907e27e1c4a",
		Name:            "Kytheon's Irregulars",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 1"},
		Activated: []ActivatedAbility{{
			Label:   "{W}{W}: Tap target creature.",
			Cost:    ManaCost("{W}{W}"),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
