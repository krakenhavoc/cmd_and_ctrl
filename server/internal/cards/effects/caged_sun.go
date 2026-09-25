package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caged Sun — Artifact {6}:
//
//	"As this artifact enters, choose a color.
//	 Creatures you control of the chosen color get +1/+1.
//	 Whenever a land's ability causes you to add one or more mana of
//	 the chosen color, add an additional one mana of that color."
//
// The anthem half is Coldsteel Heart's ETB colour prompt
// (ChooseColorAsEnters) plus the ready-made ChosenColorAnthem static.
//
// Caveat: the mana-doubling clause isn't implemented. It needs a
// replacement narrower than Mana Reflection's — ADD ONE of a specific
// chosen colour, only when the source is a LAND, not double the whole
// production — which mana_replacements.go's ManaProducedBecomes
// doesn't express (it multiplies every colour in the event
// uniformly). Only the anthem works.
func init() {
	Register(Spec{
		OracleID:     "09b895ff-e729-48d1-bfc1-ea5fd7adda6a",
		Name:         "Caged Sun",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Doubling mana of the chosen color from lands isn't implemented — only the anthem for creatures of the chosen color works."},
		AsEnters:     ChooseColorAsEnters(game.ColorForBenefit, "Caged Sun"),
		Static:       []game.StaticAbility{ChosenColorAnthem(1, 1)},
	})
}
