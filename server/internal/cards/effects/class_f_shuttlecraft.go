package effects

// Class-F Shuttlecraft — Artifact — Spacecraft {3}, 2/4 (Star Trek,
// #4973):
//
//	"{T}: Add one mana of any color.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 6+.)
//	 6+ | Flying, vigilance"
//
// Oberth-Class Starship's station shape (Station, SpacecraftAt,
// ThresholdKeywords) with Aetheric Amplifier's any-color mana ability.
// The mana ability works at any charge count; once the Spacecraft is a
// creature (6+), its {T} needs it to have been under its controller's
// control since the turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ccc91f2-4d6e-4d64-b438-3d8b4361040a",
		Name:         "Class-F Shuttlecraft",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(6, 2, 4),
			ThresholdKeywords(6, "flying", "vigilance"),
		),
	})
}
