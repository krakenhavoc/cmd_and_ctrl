package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Strings — Sorcery {1}{U}:
//
//	"You may tap or untap target permanent, then you may tap or untap
//	 another target permanent.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Two different target permanents ("another"), in announce order. For
// each still-legal one in turn (CR 608.2b) the caster chooses as it
// resolves: tap it, untap it, or leave it — the "you may" — and the
// second question is asked only once the first is answered ("then").
// Cipher is the engine's keyword (game/cipher.go, CR 702.99, #3867).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a50737d7-dd2d-40da-817c-f913f76b5011",
		Name:            "Hidden Strings",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetPermanent("target permanent, then another target permanent").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return hiddenStringsMayTapOrUntap(ctx, ids)
		},
	})
}

// hiddenStringsMayTapOrUntap asks about the first of `ids`, then the
// rest once that answer is in.
func hiddenStringsMayTapOrUntap(ctx *Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	target := ids[0]
	name := cardNameFor(ctx.Game, target)
	return PickOption{
		Question: "Hidden Strings — you may tap or untap " + name,
		Options: []game.ChoiceOption{
			{Label: "Tap it", Cards: []uuid.UUID{target}},
			{Label: "Untap it", Cards: []uuid.UUID{target}},
			{Label: "Leave it as it is", Cards: []uuid.UUID{target}},
		},
		ThenKey: hiddenStringsAnswer,
		Carry:   ids,
	}.Apply(ctx)
}

// hiddenStringsAnswer is the answer about the first target, then the next
// question about the rest, keyed (#4253): the targets still to be asked about
// ride the prompt as Carry, the one in question first. Its key is an on-disk
// identity: never renamed, never reused.
//
// Assigned in init, not in the declaration: the continuation asks the next
// question, which names this key, and a package-level initializer may not
// refer to itself (an initialization cycle).
var hiddenStringsAnswer game.OptionPickThen

func init() {
	hiddenStringsAnswer = OptionPickThen("option-pick/hidden-strings-tap-or-untap", hiddenStringsAnswered)
}

func hiddenStringsAnswered(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	target, rest := r.Carry[0], r.Carry[1:]
	var err error
	switch r.Index {
	case 0:
		err = TapTarget{Target: target}.Apply(ctx)
	case 1:
		err = UntapTarget{Target: target}.Apply(ctx)
	}
	if err != nil {
		return err
	}
	return hiddenStringsMayTapOrUntap(ctx, rest)
}
