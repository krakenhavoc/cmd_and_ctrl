package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// _____ Goblin — Creature — Goblin Guest {2}{R}, 2/2:
//
//	"When this creature enters, you may put a name sticker on it. Add {R}
//	 for each unique vowel on that sticker. (The vowels are A, E, I, O, U,
//	 and Y.)"
//
// The name sticker is the engine's keyed questions with only name stickers
// offered (ADR 0208 §4). "That sticker" is the one the first sentence put
// on, which the engine files for the resolution (game.StickerPlacedRecordKey)
// so the second sentence finds it after a restore; with none put on, there
// is no sticker and no vowel (CR 123.6e: the different vowels that appear
// on it, a sticker counted once however often a vowel repeats). The mana
// is added as the trigger resolves, so it empties with the step (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88222fd2-8316-426c-8218-64f6be5ca0f8",
		Name:         "_____ Goblin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("_____ Goblin — you may put a name sticker on it. Add {R} for each unique vowel on that sticker",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					Instruction(blankGoblinMana),
				)),
		},
	})
}

// blankGoblinMana is the second sentence: {R} for each unique vowel on the
// sticker the first one put on.
func blankGoblinMana(ctx *Context) error {
	n := uniqueVowelsOnPlacedSticker(ctx)
	if n <= 0 {
		return nil
	}
	return AddMana{Amount: ManaN(n, "R")}.Apply(ctx)
}
