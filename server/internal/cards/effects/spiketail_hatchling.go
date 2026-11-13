package effects

// Spiketail Hatchling — Creature — Drake {1}{U}, 1/1 (Prophecy, #5617):
//
//	"Flying (This creature can't be blocked except by creatures with
//	 flying or reach.)
//	 Sacrifice this creature: Counter target spell unless its controller
//	 pays {1}."
//
// Flying is a printed keyword. The ability is Diversion Unit's
// sacrifice-to-tax with any spell as the target and {1} as the tax
// (counterTheTargetSpellUnlessPaid). The Hatchling is sacrificed as the
// cost, so it is gone before the spell's controller decides; it can be
// activated the turn it arrives, since the cost has no {T}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8109aeff-1ba7-46c2-89ab-450d6ff99c28",
		Name:            "Spiketail Hatchling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target spell unless its controller pays {1}.",
			Cost:    SacrificeThis(),
			Targets: TargetSpell("target spell"),
			Effect:  counterTheTargetSpellUnlessPaid("Spiketail Hatchling", "{1}"),
		}},
	})
}
