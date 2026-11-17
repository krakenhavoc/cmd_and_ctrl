package effects

// Citanul Stalwart — Creature — Elf Druid Soldier {G}, 1/1 (The Brothers'
// War):
//
//	"{T}, Tap an untapped artifact or creature you control: Add one mana of
//	 any color."
//
// Gene Pollinator's mana ability with an artifact-or-creature filter. The
// {T} taps the Stalwart, so the other permanent is another one
// (TapAnotherUntapped); it is not the {T} symbol (CR 302.6), so one that
// arrived this turn pays it, and it does not target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce291948-125d-493d-8770-8b3aac515f9b",
		Name:         "Citanul Stalwart",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, TapOthers: TapAnotherUntapped("an untapped artifact or creature you control", Or(Artifact(), Creature())).TapOthers},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
