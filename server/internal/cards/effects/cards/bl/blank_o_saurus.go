package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "32921952-8846-418d-8689-c42dbc24a7b8",
		Name:            "_____-o-saurus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("_____-o-saurus — you may put a name sticker on it. Put a +1/+1 counter on it for each unique vowel on that sticker",
				effects.Do(
					effects.PutASticker{On: effects.StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					effects.Instruction(blankOSaurusCounters),
				)),
		},
	})
}

func blankOSaurusCounters(ctx *effects.Context) error {
	n := effects.UniqueVowelsOnPlacedSticker(ctx)
	if n <= 0 {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}
