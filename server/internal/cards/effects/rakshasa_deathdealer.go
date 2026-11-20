package effects

// Rakshasa Deathdealer — Creature — Demon {B}{G}, 2/2 (Khans of Tarkir,
// #5554):
//
//	"{B}{G}: This creature gets +2/+2 until end of turn.
//	 {B}{G}: Regenerate this creature."
//
// Two shared rows: selfPumpRow (the firebreathing shape, any number of
// times) and regenerateThisRow (a regeneration shield, CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acdc17cf-8f8c-42ee-bede-481ce00cf4bf",
		Name:         "Rakshasa Deathdealer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			selfPumpRow("Rakshasa Deathdealer", "{B}{G}", 2, 2),
			regenerateThisRow("{B}{G}"),
		},
	})
}
