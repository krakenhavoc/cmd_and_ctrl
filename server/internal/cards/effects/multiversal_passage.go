package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Multiversal Passage — Land (Marvel's Spider-Man promos, #4460):
//
//	"As this land enters, choose a basic land type. Then you may pay 2
//	 life. If you don't, it enters tapped.
//	 This land is the chosen type."
//
// The type is Convincing Mirage's as-enters choice (CR 614.12), stored on
// the permanent (Card.ChosenOption), and the life is the shockland
// replacement (EntersTappedUnlessYouPayLife): neither answer changes the
// other, so the engine asking the life first (it is settled before the land
// moves) changes nothing a player can see.
//
// "This land is the chosen type" is one layer-4 static per basic land
// type, each gated on that type being the chosen one, setting the land's
// land types to it; from the type comes the land's mana ability (CR
// 305.6). It is not written as CR 305.7's ability-removing set: the
// abilities the land's own text gives it are this choice and this static,
// and an ability another effect grants it is kept either way, so the
// removal would take nothing, and taken off a land by itself it would take
// the static that makes it the type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17770eeb-ecbf-4510-af2d-acececfbe146",
		Name:         "Multiversal Passage",
		Completeness: CompletenessFull,
		AsEnters:     ChooseBasicLandTypeAsEnters("Multiversal Passage"),
		Replacements: []game.ReplacementEffect{
			EntersTappedUnlessYouPayLife("Multiversal Passage", shocklandLifeCost),
		},
		Static: thisLandIsTheChosenType(),
	})
}

// thisLandIsTheChosenType is "This land is the chosen type": one gated
// layer-4 static per basic land type over the source itself.
func thisLandIsTheChosenType() []game.StaticAbility {
	out := make([]game.StaticAbility, 0, len(game.BasicLandTypes))
	for _, t := range game.BasicLandTypes {
		landType := t
		out = append(out, StaticWhenChosen(landType, game.StaticAbility{
			Label:     "This land is the chosen type.",
			Layer:     game.Layer4Type,
			AppliesTo: thisPermanent,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.SetLandSubtypes([]string{landType})
			},
		}))
	}
	return out
}
