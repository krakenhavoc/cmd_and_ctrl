package effects

// Initiates of the Ebon Hand — Creature — Cleric {B}, 1/1 (Fallen Empires,
// #5745):
//
//	"{1}: Add {B}. If this ability has been activated four or more times
//	 this turn, sacrifice this creature at the beginning of the next end
//	 step."
//
// Farrelite Priest's ability in black: a {1} filter with no tap, whose
// activation-count rider (mana_activation_count_rider.go) sacrifices the
// Initiates at the beginning of the next end step from the fourth
// activation of this ability on this object this turn. The auto-tapper
// does not spend it; the player activates it from the board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bf359ef-669f-41dc-8ab5-ba8e58b2f398",
		Name:         "Initiates of the Ebon Hand",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Mana: "{1}"},
			Produced: "{B}",
			Label:    initiatesOfTheEbonHandManaLabel,
			Rider: SacrificeAtNextEndStepIfActivatedRider(4, initiatesOfTheEbonHandManaLabel,
				"Initiates of the Ebon Hand — sacrifice this creature"),
		}},
	})
}

// initiatesOfTheEbonHandManaLabel is the mana ability's label, and so the
// key its activations are counted under.
const initiatesOfTheEbonHandManaLabel = "{1}: Add {B}. If this ability has been activated four or more times this turn, sacrifice this creature at the beginning of the next end step"
