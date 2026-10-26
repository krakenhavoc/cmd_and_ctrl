package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// _____ _____ Rocketship — Artifact — Vehicle {4}, 2/3:
//
//	"Flying
//	 When this Vehicle enters, you may put up to two name stickers on it.
//	 Whenever this Vehicle attacks, choose a letter. This Vehicle gets
//	 +1/+1 until end of turn for each name sticker on it that begins with
//	 the chosen letter.
//	 Crew 2"
//
// "Up to two" is two PutASticker steps in order: each is "you may", and a
// player who declines the first and takes the second ends where one who
// took one sticker would (CR 123.3: each is its own instruction). Crew is
// the engine's, from the card's printed text.
//
// "Choose a letter" is a keyed option pick of the 26 letters, so a table
// waiting on it is a restore point; any letter may be chosen, also one no
// name sticker begins with (CR 123.6d: a lowercase letter and its uppercase
// equivalent are the same letter). The count is read when the letter is
// chosen, from the Vehicle while it stands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1009f979-b1fa-4427-bf9e-2a533a4ca19b",
		Name:            "_____ _____ Rocketship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("_____ _____ Rocketship — you may put up to two name stickers on it",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
				)),
			WheneverThisAttacks("_____ _____ Rocketship — choose a letter; it gets +1/+1 until end of turn for each name sticker on it that begins with that letter",
				Do(Instruction(blankRocketshipChooseLetter))),
		},
	})
}

// blankRocketshipLetterThen is the answer to "choose a letter": the
// Vehicle (Carry) gets +1/+1 for each name sticker on it beginning with
// the letter the option names.
var blankRocketshipLetterThen = OptionPickThen("option-pick/blank-blank-rocketship-letter", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	letter, _ := firstRune(r.Option.Label)
	vehicle := r.Carry[0]
	c, ok := ctx.Game.LookupCardForEffect(vehicle)
	if !ok || !onBattlefield(ctx.Game, vehicle) {
		return nil
	}
	n := nameStickerCount(c, func(text string) bool { return game.StickerBeginsWith(text, letter) })
	return BoostUntilEOT{Target: vehicle, Power: n, Toughness: n, Label: "_____ _____ Rocketship"}.Apply(ctx)
})

func firstRune(s string) (rune, bool) {
	for _, r := range s {
		return r, true
	}
	return 0, false
}

// blankRocketshipChooseLetter asks the attacking Vehicle's controller for a
// letter, A to Z.
func blankRocketshipChooseLetter(ctx *Context) error {
	if ctx.isNewSourceObject(ctx.Source()) || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	options := make([]game.ChoiceOption, 0, 26)
	for r := 'A'; r <= 'Z'; r++ {
		options = append(options, game.ChoiceOption{Label: string(r)})
	}
	return PickOption{
		Question: "_____ _____ Rocketship — choose a letter",
		Options:  options,
		ThenKey:  blankRocketshipLetterThen,
		Carry:    []uuid.UUID{ctx.Source()},
	}.Apply(ctx)
}
