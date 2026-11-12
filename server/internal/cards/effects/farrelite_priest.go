package effects

// Farrelite Priest — Creature — Human Cleric {1}{W}{W}, 1/3 (Fallen
// Empires, #5745):
//
//	"{1}: Add {W}. If this ability has been activated four or more times
//	 this turn, sacrifice this creature at the beginning of the next end
//	 step."
//
// A filter paid with {1} and no tap, so it can be activated any number of
// times. The sentence after "Add {W}" is the activation-count rider
// (mana_activation_count_rider.go): from the fourth activation of this
// ability on this object this turn, a delayed trigger sacrifices the
// Priest at the beginning of the next end step. The auto-tapper does not
// spend it; the player activates it from the board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ee18e5f-3d12-489e-b337-29208caaed7a",
		Name:         "Farrelite Priest",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Mana: "{1}"},
			Produced: "{W}",
			Label:    farrelitePriestManaLabel,
			Rider: SacrificeAtNextEndStepIfActivatedRider(4, farrelitePriestManaLabel,
				"Farrelite Priest — sacrifice this creature"),
		}},
	})
}

// farrelitePriestManaLabel is the mana ability's label, and so the key
// its activations are counted under.
const farrelitePriestManaLabel = "{1}: Add {W}. If this ability has been activated four or more times this turn, sacrifice this creature at the beginning of the next end step"
