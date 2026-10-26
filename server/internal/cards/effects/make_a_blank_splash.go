package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Make a _____ Splash — Enchantment {2}{U}:
//
//	"Flash
//	 When this enchantment enters, you may put a name sticker on it.
//	 Whenever you put a sticker on this enchantment, tap up to X target
//	 creatures, where X is the number of u's in name stickers on this
//	 enchantment."
//
// The sticker the enters trigger puts on triggers the second ability too,
// as does every later one. X sets how many targets the ability may have,
// so it is read as the ability is put on the stack (CR 603.3d), from the
// words of the name stickers on the enchantment that the trigger carries
// (game.Event.Words): a trigger context, so a restored table announces the
// same clause. CR 123.6d: a lowercase letter and its uppercase equivalent
// are the same letter. With no u at all it is "up to 0": the ability goes
// on the stack with no target and taps nothing.
//
// No simplification.
func init() {
	tap := On(game.EventStickerPlaced, aStickerWentOnThis,
		"Make a _____ Splash — tap up to X target creatures, X the number of u's in name stickers on it",
		makeABlankSplashTap)
	tap.TargetsFrom = makeABlankSplashTargets
	Register(Spec{
		OracleID:        "fe2601a4-cbd5-4866-92a2-5586f737c9af",
		Name:            "Make a _____ Splash",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Make a _____ Splash — you may put a name sticker on it",
				Do(PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}})),
			tap,
		},
	})
}

// makeABlankSplashTargets is "up to X target creatures", X the u's in the
// name stickers on the enchantment once the sticker that triggered this
// went on. With X = 0 there is nothing to target.
func makeABlankSplashTargets(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	x := lettersOnWords(tc.Event.Words, 'u')
	if x <= 0 {
		return nil
	}
	return TargetCreature(fmt.Sprintf("up to %d target creatures (X = %d)", x, x)).WithCount(0, x)
}

func makeABlankSplashTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCardIDs(ctx) {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
