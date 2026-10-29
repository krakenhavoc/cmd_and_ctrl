package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wolf in _____ Clothing — Creature — Wolf Guest {3}{B}, 2/3:
//
//	"When this creature enters, you may put a name sticker on it. When you
//	 do, up to X target creatures each get -1/-1 until end of turn, where X
//	 is the number of unique vowels on that sticker. (The vowels are A, E,
//	 I, O, U, and Y.)"
//
// "That sticker" is the one the first sentence put on (ADR 0208, CR 123.6e).
// "When you do" is a reflexive trigger (CR 603.12), created only when a
// sticker went on: it goes on the stack above the enters trigger and its
// targets are chosen then, "up to X" with X carried as the trigger's
// Params.Amount (a registered body, so a table waiting on it is a restore
// point). A sticker with no vowel at all makes "up to 0": the trigger would
// shrink nothing, so it is not put on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58c9454b-7027-48f4-9e17-310e27b916d1",
		Name:         "Wolf in _____ Clothing",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wolf in _____ Clothing — you may put a name sticker on it. When you do, up to X target creatures each get -1/-1 until end of turn",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					Instruction(wolfInBlankClothingWhenYouDo),
				)),
		},
	})
}

// wolfInBlankClothingWhenYouDo is "When you do": a sticker went on, so the
// reflexive trigger is created, X the unique vowels on it.
func wolfInBlankClothingWhenYouDo(ctx *Context) error {
	x := uniqueVowelsOnPlacedSticker(ctx)
	if x <= 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Wolf in _____ Clothing — up to X target creatures each get -1/-1 until end of turn",
		Body:   wolfInBlankClothingShrinkBody,
		Params: game.EffectParams{Amount: x},
	}.Apply(ctx)
}

// wolfInBlankClothingShrinkBody is the reflexive half: each target still
// legal gets -1/-1 until end of turn. The clause is "up to X target
// creatures", X from Params.Amount.
var wolfInBlankClothingShrinkBody = game.ReflexiveBody("wolf-in-blank-clothing/shrink", simpleBody(wolfInBlankClothingShrink),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		x := max(p.Amount, 1)
		return TargetCreature(fmt.Sprintf("up to %d target creatures", x)).WithCount(0, x)
	})

func wolfInBlankClothingShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCardIDs(ctx) {
		if err := (BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Wolf in _____ Clothing"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
