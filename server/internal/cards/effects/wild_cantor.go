package effects

// Wild Cantor — Creature — Human Druid {R/G}, 1/1 (Guildpact, #5599):
//
//	"({R/G} can be paid with either {R} or {G}.)
//	 Sacrifice this creature: Add one mana of any color."
//
// Goldhound's mana ability without the {T}: a mana ability (CR 605.1a),
// so it resolves at once, can be activated while paying a cost, and a
// creature that just arrived may use it (no {T} symbol, CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24472823-d33c-4d96-add1-233116f6f4f7",
		Name:         "Wild Cantor",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Sacrifice this creature: Add one mana of any color",
		}},
	})
}
