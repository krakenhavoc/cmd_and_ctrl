package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adaptive Automaton — Artifact Creature — Construct, {3}, 2/2:
//
//	"As this creature enters, choose a creature type.
//	 This creature is the chosen type in addition to its other types.
//	 Other creatures you control of the chosen type get +1/+1."
//
// The colourless lord: it goes in every tribal deck because it needs
// no colour and picks its own tribe. It is also the one card in the
// sprint that exercises the CR 613 layer ordering for real — the
// self-type-add is layer 4 and the anthem is layer 7c, and the layers
// run in that order, which is why an Adaptive Automaton naming
// Construct correctly does NOT pump itself ("other") while a second
// Automaton naming Construct does pump the first.
//
// The type-add is `IsAlsoTheChosenType` (tribal.go), shared with
// Roaming Throne, which prints the same sentence. It is an append
// with a membership check rather than an unconditional one: a
// permanent that already has the type (naming Construct, or a
// changeling under a Maskwood Nexus) must not end up with it twice,
// because the wire type line is rebuilt from this slice and would
// print "Construct Construct".
func init() {
	Register(Spec{
		OracleID: "53c730c6-2f8c-4af8-b400-b9d573a71e60",
		Name:     "Adaptive Automaton",
		AsEnters: ChooseCreatureTypeAsEnters("Adaptive Automaton"),
		Static: []game.StaticAbility{
			IsAlsoTheChosenType(),
			TribalAnthem(TribeFilter{Chosen: true, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
