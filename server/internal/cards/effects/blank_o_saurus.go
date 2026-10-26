package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// _____-o-saurus — Creature — Alien Dinosaur {4}{G}{G}, 3/3:
//
//	"Trample
//	 When this creature enters, you may put a name sticker on it. Put a
//	 +1/+1 counter on it for each unique vowel on that sticker. (The
//	 vowels are A, E, I, O, U, and Y.)"
//
// See _____ Goblin: "that sticker" is the one the first sentence put on
// (ADR 0208, CR 123.6e). The counters go on the creature as one instruction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "32921952-8846-418d-8689-c42dbc24a7b8",
		Name:            "_____-o-saurus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("_____-o-saurus — you may put a name sticker on it. Put a +1/+1 counter on it for each unique vowel on that sticker",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					Instruction(blankOSaurusCounters),
				)),
		},
	})
}

func blankOSaurusCounters(ctx *Context) error {
	n := uniqueVowelsOnPlacedSticker(ctx)
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}
