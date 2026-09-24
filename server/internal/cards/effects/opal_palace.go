package effects

// Opal Palace — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color in your commander's color
//	 identity. If you spend this mana to cast your commander, it
//	 enters with a number of additional +1/+1 counters on it equal to
//	 the number of times it's been cast from the command zone this
//	 game."
//
// Both mana abilities are real: the colourless tap, and the {1}, {T}
// filter ability that narrows to the controller's commander's colour
// identity exactly as Arcane Signet does (NarrowToCommanderIdentity),
// with the Signet cycle's {1} mana cost added on top of the tap.
//
// S49 sandbox simplification: the bonus +1/+1 counters aren't
// implemented. A mana-spend rider can make a creature enter with a
// FIXED number of extra counters (Biophagus), but this one scales
// with how many times your commander has already been cast from the
// command zone this game — a count the rider machinery has no slot
// for. Casting your commander with this land's mana still works; it
// just enters without the bonus counters.
func init() {
	Register(Spec{
		OracleID:     "aa6723a2-75da-49f5-a1ba-cbfa82c55301",
		Name:         "Opal Palace",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"When you spend this land's colored mana to cast your commander, it doesn't get the extra +1/+1 counters for past casts from the command zone — that count has no cost/rider shape yet."},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:                      ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced:                  "{W|U|B|R|G}",
				Label:                     "{1}, {T}: Add one mana of any color in your commander's color identity",
				NarrowToCommanderIdentity: true,
			},
		},
	})
}
