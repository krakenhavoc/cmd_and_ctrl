package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_shared.go — the keyed "you may" answers many cards share (#4291).
//
// A MayChoice whose answer is a registered key (MayChoice.Then) holds no
// closure, so a table waiting on it is a restore point. Most callers have
// an answer of their own, a key per card. These are the consequences that
// recur word for word, so they share one key, and the card's own number or
// name rides the prompt as MayChoice.Word (and the life a branch is
// declared to cost as LifeCost). Each key is an on-disk identity: never
// renamed, never reused, ledgered "confirm <key>".

// mayDrawCards is "you may draw N cards" for the player asked (Word is N).
// Bringer of the Blue Dawn, Spiritual Focus, Heartwood Storyteller,
// Tataru Taru.
var mayDrawCards = MayYes("confirm/draw-cards", func(ctx *Context, r game.ConfirmAnswer) error {
	return DrawCards{Player: r.Chooser, N: wordInt(r)}.Apply(ctx)
})

// mayGainLife is "you may gain N life" for the player asked (Word is N).
// Deathgreeter.
var mayGainLife = MayYes("confirm/gain-life", func(ctx *Context, r game.ConfirmAnswer) error {
	return GainLife{Player: r.Chooser, Amount: wordInt(r)}.Apply(ctx)
})

// mayCreateTokens is "you may create N <token>" for the player asked,
// the Word being tokenWord(N, template). Herd Baloth, Scurry Oak, Darien,
// King of Kjeldor.
var mayCreateTokens = MayYes("confirm/create-tokens", func(ctx *Context, r game.ConfirmAnswer) error {
	n, template := parseTokenWord(r.Word)
	if n <= 0 || template == "" {
		return nil
	}
	return CreateToken{Controller: r.Chooser, Template: TokenCard(template), N: n}.Apply(ctx)
})

// tokenWord is the Word for mayCreateTokens: "<n> <template>".
func tokenWord(n int, template string) string {
	return strconv.Itoa(n) + " " + template
}

func parseTokenWord(w string) (int, string) {
	count, template, ok := strings.Cut(w, " ")
	if !ok {
		return 0, ""
	}
	n, err := strconv.Atoi(count)
	if err != nil {
		return 0, ""
	}
	return n, template
}

// mayCopyThisSpell is "you may copy this spell, and may choose a new
// target for the copy", copied under the control of the player asked. The
// Chain of ... cycle and Sevinne's Reclamation.
var mayCopyThisSpell = MayYes("confirm/copy-this-spell", func(ctx *Context, r game.ConfirmAnswer) error {
	return copyThisSpellFor(r.Chooser)(ctx)
})

// mayPayLifeDraw is "you may pay N life. If you do, draw M cards" for the
// player asked: N is the life the prompt declares (LifeCost) and the Word
// is M. The payment is re-checked when the answer arrives (CR 119.4): life
// moves between the question and the answer. Call of the Ring, Erebos,
// Bleak-Hearted, Crossway Troublemakers, Gix, Yawgmoth Praetor.
var mayPayLifeDraw = MayYes("confirm/pay-life-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	return payLifeThenDrawFor(ctx, r.Chooser, r.LifeCost, wordInt(r))
})

// maySearchNamedLibraryOrGraveyard is "you may search your library and/or
// graveyard for a card named X, reveal it and put it into your hand"
// (Word is X). Ajani's Aid and the creatures that search for a partner
// card by name.
var maySearchNamedLibraryOrGraveyard = MayYes("confirm/search-named-library-or-graveyard", func(ctx *Context, r game.ConfirmAnswer) error {
	named := r.Word
	return SearchLibrary{
		Player:        r.Chooser,
		Predicate:     func(c game.Card) bool { return c.Name == named },
		Dest:          game.ZoneHand,
		Limit:         1,
		Reveal:        true,
		Shuffle:       true,
		AlsoGraveyard: true,
		Reason:        "Choose a card named " + named + " to put into your hand",
	}.Apply(ctx)
})

// carryAt is the i-th ID the card put on the prompt (r.Carry), uuid.Nil when
// there is none.
func carryAt(r game.ConfirmAnswer, i int) uuid.UUID {
	if i < len(r.Carry) {
		return r.Carry[i]
	}
	return uuid.Nil
}

// joinCarry is two lists of IDs as one Carry, with uuid.Nil between them:
// the prompt holds one list, and a recursive "ask each player in turn"
// answer needs two (those who have said yes, and those still to ask).
func joinCarry(before, after []uuid.UUID) []uuid.UUID {
	out := append([]uuid.UUID(nil), before...)
	out = append(out, uuid.Nil)
	return append(out, after...)
}

// splitCarry is joinCarry's inverse, on the card's own Carry.
func splitCarry(r game.ConfirmAnswer) (before, after []uuid.UUID) {
	for i, id := range r.Carry {
		if id == uuid.Nil {
			return append([]uuid.UUID(nil), r.Carry[:i]...), append([]uuid.UUID(nil), r.Carry[i+1:]...)
		}
	}
	return append([]uuid.UUID(nil), r.Carry...), nil
}
