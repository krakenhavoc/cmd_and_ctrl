package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// _____ Balls of Fire — Enchantment {3}{R}:
//
//	"When this enchantment enters, you may put a name sticker on it.
//	 Whenever you put a sticker on this enchantment, it deals damage equal
//	 to the number of o's in name stickers on this enchantment to any
//	 target."
//
// The sticker the enters trigger puts on triggers the second ability too
// (it is a sticker put on this enchantment); so does every later one,
// whichever effect puts it on. The count is read as the ability resolves
// (CR 123.6d: a lowercase letter and its uppercase equivalent are the same
// letter), from the enchantment while it stands and from the words the
// trigger carried (game.Event.Words) once it has left (CR 608.2h).
//
// No simplification.
func init() {
	damage := On(game.EventStickerPlaced, aStickerWentOnThis,
		"_____ Balls of Fire — it deals damage equal to the number of o's in name stickers on it to any target",
		blankBallsOfFireDamage)
	damage.Targets = TargetAny()
	Register(Spec{
		OracleID:     "82c2a58b-0ee2-454b-9ddd-670c1e788166",
		Name:         "_____ Balls of Fire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("_____ Balls of Fire — you may put a name sticker on it",
				Do(PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}})),
			damage,
		},
	})
}

// aStickerWentOnThis is "whenever you put a sticker on this <permanent>":
// the source's controller put a sticker on the source itself.
func aStickerWentOnThis(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller && ev.CardID == source.InstanceID
}

// nameStickerWordsNow is the name-sticker words on the trigger's source as
// the ability resolves: the live object while it stands (CR 123.5: a
// sticker stays on a card that changes public zones), else the words the
// trigger was put on the stack with.
func nameStickerWordsNow(ctx *Context) []string {
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok && !ctx.isNewSourceObject(ctx.Source()) {
		if words := nameStickersOn(c); len(words) > 0 {
			return words
		}
	}
	if ctx.Item != nil && ctx.Item.Trigger != nil {
		return append([]string(nil), ctx.Item.Trigger.Event.Words...)
	}
	return nil
}

// lettersOnWords is the number of times `letter` appears in the words.
func lettersOnWords(words []string, letter rune) int {
	n := 0
	for _, w := range words {
		n += game.StickerLetterCount(w, letter)
	}
	return n
}

func blankBallsOfFireDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	n := lettersOnWords(nameStickerWordsNow(ctx), 'o')
	if n <= 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
}
