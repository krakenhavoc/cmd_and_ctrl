package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wizards of the _____ — Creature — Human Wizard Performer {2}{U}{U}, 3/1:
//
//	"When this creature enters, you may put a name sticker on it, then
//	 look at the top X cards of your library, where X is the number of
//	 unique vowels on that sticker. Put one of those cards into your hand
//	 and the rest on the bottom of your library in any order. (The vowels
//	 are A, E, I, O, U, and Y.)"
//
// See _____ Goblin: "that sticker" is the one the first sentence put on
// (ADR 0208, CR 123.6e). With none put on, X is 0 and nothing is looked
// at. The look is Impulse's (LookAtTopThenTakeOneRestOnBottom): only the
// looker sees the cards, one goes to the hand, the rest to the bottom in
// an order the player chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "786b0b50-adc7-4d58-b578-1e80680b89bd",
		Name:         "Wizards of the _____",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wizards of the _____ — you may put a name sticker on it, then look at the top X cards of your library",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					Instruction(wizardsOfTheBlankLook),
				)),
		},
	})
}

func wizardsOfTheBlankLook(ctx *Context) error {
	x := uniqueVowelsOnPlacedSticker(ctx)
	if x <= 0 {
		return nil
	}
	return LookAtTopThenTakeOneRestOnBottom(x, "Wizards of the _____ — put one of them into your hand")(ctx.Game, ctx.Item)
}
