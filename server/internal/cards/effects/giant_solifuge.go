package effects

// Giant Solifuge — Creature — Insect {2}{R/G}{R/G}, 4/1 (Guildpact,
// #5599):
//
//	"({R/G} can be paid with either {R} or {G}.)
//	 Trample; haste; shroud (This creature can't be the target of spells
//	 or abilities.)"
//
// Three printed keywords, each one the engine already enforces: trample
// in combat damage, haste for attacking and {T}, shroud at the targeting
// choke point. The hybrid symbols are printed mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c590c1bd-430c-4515-9343-ba818c02bc5f",
		Name:            "Giant Solifuge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste", "shroud"},
	})
}
