package effects

// energy_mana.go — ADR 0129 §5: the shared rows of the mana abilities
// that pay energy (Aether Hub, Solar Transformer, Servant of the Conduit).

// colorlessTapRow is "{T}: Add {C}."
func colorlessTapRow() ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{C}",
		Label:    "{T}: Add {C}.",
	}
}

// anyColorForEnergyRow is "{T}, Pay N {E}: Add one mana of any color."
func anyColorForEnergyRow(energy int) ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true, Energy: energy},
		Produced: "{W|U|B|R|G}",
		Label:    "{T}, Pay " + EnergySymbols(energy) + ": Add one mana of any color.",
	}
}

// anyColorForEnergyRows is the {C} row followed by the energy row, the
// pair Aether Hub and Solar Transformer print.
func anyColorForEnergyRows(energy int) []ManaAbility {
	return []ManaAbility{colorlessTapRow(), anyColorForEnergyRow(energy)}
}
