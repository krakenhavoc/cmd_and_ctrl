package effects

// Aron, Benalia's Ruin — Legendary Creature — Phyrexian Human {W}{W}{B},
// 3/3:
//
//	"Menace
//	 {W}{B}, {T}, Sacrifice another creature: Put a +1/+1 counter on each
//	 creature you control."
//
// "Another" is object identity (SacrificeAnotherN). The counters go on
// the creatures you control as the ability resolves, the sacrificed
// creature no longer among them (b11PutCounterOnEachCreatureYouControl).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8b06606-9b1f-4e72-a280-12a6e912bb50",
		Name:            "Aron, Benalia's Ruin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:  "{W}{B}, {T}, Sacrifice another creature: Put a +1/+1 counter on each creature you control.",
			Cost:   Plus(ManaCost("{W}{B}"), TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: b11PutCounterOnEachCreatureYouControl,
		}},
	})
}
