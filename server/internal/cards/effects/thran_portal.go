package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thran Portal — Land — Gate (Dominaria United, #5661):
//
//	"This land enters tapped unless you control two or fewer other
//	 lands.
//	 As this land enters, choose a basic land type.
//	 This land is the chosen type in addition to its other types.
//	 Mana abilities of this land cost an additional 1 life to
//	 activate."
//
// The first line is the fastlands' replacement, counted before the land
// is on the battlefield. The choice is Convincing Mirage's as-enters
// one (CR 614.12), stored on the permanent (Card.ChosenOption). "In
// addition to its other types" adds the chosen type in layer 4 and
// keeps Gate and the land's text (CR 305.7's last sentence), unlike
// Multiversal Passage's "is the chosen type"; the mana ability comes
// from the type (CR 305.6), so no row is declared here.
//
// That is why the last line is an engine seam rather than a "Pay 1
// life" on a row: the ability it prices is not printed on the card.
// ManaAbilitiesOfThisPermanentCostLife charges every mana ability the
// land has, a granted one included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "926ce6a2-7bdd-4380-ac65-bc902ba0c284",
		Name:         "Thran Portal",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(otherLandsAtMost(2))},
		AsEnters:     ChooseBasicLandTypeAsEnters("Thran Portal"),
		Static: append(thisLandIsTheChosenTypeInAddition(),
			ManaAbilitiesOfThisPermanentCostLife(1)),
	})
}

// thisLandIsTheChosenTypeInAddition is "This land is the chosen type in
// addition to its other types": one gated layer-4 static per basic land
// type over the source itself, adding the type and removing nothing.
func thisLandIsTheChosenTypeInAddition() []game.StaticAbility {
	out := make([]game.StaticAbility, 0, len(game.BasicLandTypes))
	for _, t := range game.BasicLandTypes {
		landType := t
		out = append(out, StaticWhenChosen(landType, game.StaticAbility{
			Label:     "This land is the chosen type in addition to its other types.",
			Layer:     game.Layer4Type,
			AppliesTo: thisPermanent,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Subtypes, landType) {
					c.Subtypes = append(c.Subtypes, landType)
				}
			},
		}))
	}
	return out
}
