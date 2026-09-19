package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patchwork Banner — Artifact {3}:
//
//	"As this artifact enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1.
//	 {T}: Add one mana of any color."
//
// Vanquisher's Banner's cheaper cousin: the same CR 614.12 creature
// type named on entry and the same chosen-tribe anthem, with a
// five-colour mana ability instead of the draw trigger.
//
// The anthem is `TribeFilter{Chosen: true, YoursOnly: true}` — "of the
// chosen type" and "you control", both printed. It reads effective
// subtypes, so a changeling is every type and counts (CR 702.73a),
// and until the controller answers the prompt the named type is empty
// and the filter matches nothing. An empty type read as "everything"
// would be the dangerous direction, which is why nothing in the
// catalogue writes one.
//
// The mana ability is unrestricted "any color" — the printed text
// says nothing about the commander's identity, so all five colours
// are offered and NarrowToCommanderIdentity stays off. It is not
// gated on the chosen type either: Patchwork Banner taps for mana on
// turn three whatever it named.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fb00dbe-1f82-4ba6-b18c-97e816d10d3a",
		Name:         "Patchwork Banner",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Patchwork Banner"),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 1, 1),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
