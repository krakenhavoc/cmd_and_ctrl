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
// The mana half is a CR 605.1b triggered mana ability (#763), the same
// family as Wild Growth and Mana Flare rather than a Mana-Reflection-
// style replacement: ManaProducedBecomes multiplies every colour a
// production made uniformly, which is the wrong shape for "add ONE
// more, only of the chosen colour, only from a land."
// WheneverYouTapALandForManaOfTheChosenColor (mana_triggers.go) is
// built for exactly that: it fires only when a land YOU control
// produces at least one mana of the chosen colour, and always adds
// exactly one more of it — a dual land tapped for the chosen colour
// gets the bonus, the same dual tapped for its other colour does not,
// and a land making two of the chosen colour still gets only one
// extra. Before a colour is chosen it adds nothing.
//
// "A land's ability" vs. "a land tapped for mana": the engine's three
// triggered-mana firing sites all gate on the mana ability having a
// TAP cost (CR 605.1b reads "tapped for mana"), so a hypothetical land
// whose mana ability didn't tap would not trigger this. Every
// catalogued land's mana ability has a Tap cost as of S44, so the two
// readings coincide for every land in this catalog today and the gap
// is not observable at the table.
func init() {
	Register(Spec{
		OracleID:     "09b895ff-e729-48d1-bfc1-ea5fd7adda6a",
		Name:         "Caged Sun",
		Completeness: CompletenessFull,
		AsEnters:     ChooseColorAsEnters(game.ColorForBenefit, "Caged Sun"),
		Static:       []game.StaticAbility{ChosenColorAnthem(1, 1)},
		ManaTriggers: []game.ManaTrigger{
			WheneverYouTapALandForManaOfTheChosenColor("Caged Sun — add an additional one mana of the chosen color"),
		},
	})
}
