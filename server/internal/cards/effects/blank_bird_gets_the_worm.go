package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// _____ Bird Gets the Worm — Creature — Bird Guest {3}{W}{W}, 2/4:
//
//	"Flying
//	 When this creature enters, you may put a name sticker on it. You gain
//	 X life, where X is the number of unique vowels on that sticker. (The
//	 vowels are A, E, I, O, U, and Y.)"
//
// See _____ Goblin: "that sticker" is the one the first sentence put on
// (ADR 0208, CR 123.6e). With none put on, X is 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2ce375c8-7c37-40b3-95cf-e0aa958bb317",
		Name:            "_____ Bird Gets the Worm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("_____ Bird Gets the Worm — you may put a name sticker on it. You gain X life, X the unique vowels on that sticker",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					Instruction(blankBirdLife),
				)),
		},
	})
}

func blankBirdLife(ctx *Context) error {
	return GainLife{Amount: uniqueVowelsOnPlacedSticker(ctx)}.Apply(ctx)
}
