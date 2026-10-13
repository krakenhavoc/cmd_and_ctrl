package effects

// Faithful Squire // Kaiso, Memory of Loyalty — {1}{W}{W} Creature —
// Human Soldier 2/2 // Legendary Creature — Spirit 3/4, a flip card
// (#3216, ADR 0151):
//
//	Top:     "Whenever you cast a Spirit or Arcane spell, you may put a
//	          ki counter on this creature.
//	          At the beginning of the end step, if there are two or more
//	          ki counters on this creature, you may flip it."
//	Flipped: "Flying
//	          Remove a ki counter from Kaiso: Prevent all damage that
//	          would be dealt to target creature this turn."
//
// Kaiso's flying is its own half's keyword (ADR 0151 decision 3), so
// the Squire does not fly and the flip gives the permanent flying. The
// shield is the shared "prevent all damage that would be dealt to target
// creature this turn" row. Kaiso is white and {1}{W}{W} (CR 710.1c).
//
// No simplification.
const faithfulSquireOracleID = "2358ffc2-6663-4ef3-b3a4-b036a4733ac6"

func init() {
	Register(Spec{
		OracleID:     faithfulSquireOracleID,
		Name:         "Faithful Squire",
		Completeness: CompletenessFull,
		Triggered:    kiFlipTriggers("Faithful Squire"),
	})
	Register(Spec{
		OracleID:        faithfulSquireOracleID + "#1",
		Name:            "Kaiso, Memory of Loyalty",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			shieldTargetRow("Remove a ki counter from Kaiso: Prevent all damage that would be dealt to target creature this turn.",
				removeAKiCounter(), TargetCreature("target creature")),
		},
	})
}
