package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_or_outside.go — "You may reveal a <kind> card you own from
// outside the game or choose a face-up <kind> card you own in exile.
// Put that card into your hand." (Coax from the Blind Eternities; Karn,
// the Great Creator's −2.)
//
// One choice with two places to look. It is asked in order: first the
// matching face-up cards the player owns in exile, with "none" meaning
// "look outside the game instead", then, if they took none, the ordinary
// pick from outside the game (TakeFromOutside, revealed). The player can
// reach every answer the printed choice offers and no other: a card from
// exile, a card from outside, or nothing. With nothing in exile the first
// question is not asked at all.
//
// A card chosen in exile is put into its owner's hand from exile; it is
// face up, so it is already public and is not revealed again.
//
// The pick from outside the game is keyed (#4256, outside_game.go), so a
// table waiting on it is a restore point, and so is the exile pick before
// it (#4264): a keyed choose_cards pick (registerExileOrOutside) whose
// continuation puts the chosen card into the hand, or, on "none", asks
// the outside pick with the card's word, which rides the prompt as its
// string carry (ChooseCardsPrompt.CarryWord). The keyed pick has no
// live-zone re-check of its own, so the continuation checks the card is
// still a face-up card the player owns in exile.

// TakeFromExileOrOutside is the shape described above.
type TakeFromExileOrOutside struct {
	// Pick is the card's registered outside pick (registerOutsideTake,
	// revealing), whose filter is the kind of card ("an Eldrazi card",
	// "an artifact card") in exile too. Required.
	Pick outsideTake
	// Exile is the card's registered exile pick (registerExileOrOutside),
	// the first prompt's continuation. Required.
	Exile exileTake
	// Word is the word of the card's text the filter reads, if any.
	Word string
	// ExileQuestion heads the first prompt. The second's header is the
	// exile pick's (registerExileOrOutside).
	ExileQuestion string
}

// exileTake is a registered exile pick: the first prompt of a
// TakeFromExileOrOutside, keyed.
type exileTake struct {
	then game.ChooseCardsThen
	// question heads the outside pick the "none" answer asks.
	question string
}

// registerExileOrOutside registers the exile half of a "from outside the
// game or in exile" card under `key` (#4264), whose "none" asks `outside`
// with `question`. Call it once, from a package-level var. The key is an
// on-disk identity, ledgered "choose <key>".
func registerExileOrOutside(key string, outside outsideTake, question string) exileTake {
	return exileTake{question: question, then: game.RegisterChooseCardsThen(key, func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered {
			return nil
		}
		if len(r.Picked) == 1 {
			// "Put that card into your hand", only if it is still the
			// face-up card the player owns in exile.
			if !stillFaceUpInExile(g, r.Chooser, r.Picked[0]) {
				return nil
			}
			return nothingIfGone(g.BounceToHandForEffect(r.Picked[0]))
		}
		takeFromOutside(g, r.Chooser, r.Source, TakeFromOutside{
			Player:   r.Chooser,
			Pick:     outside,
			Word:     r.Word,
			Question: question,
		})
		return nil
	})}
}

// stillFaceUpInExile reports whether `id` is a face-up card `owner` owns
// in exile now.
func stillFaceUpInExile(g *game.Game, owner, id uuid.UUID) bool {
	ex := g.Exile
	if ex == nil {
		return false
	}
	for _, c := range ex.Cards {
		if c.InstanceID == id {
			return c.Owner == owner && !c.FaceDown
		}
	}
	return false
}

func (t TakeFromExileOrOutside) Apply(ctx *Context) error {
	me := ctx.Controller()
	var exiled []uuid.UUID
	if ex := ctx.Game.Exile; ex != nil {
		for _, c := range ex.Cards {
			if c.Owner == me && !c.FaceDown && t.Pick.matches(c, t.Word) {
				exiled = append(exiled, c.InstanceID)
			}
		}
	}
	if len(exiled) == 0 {
		return TakeFromOutside{Player: me, Pick: t.Pick, Word: t.Word, Question: t.Exile.question}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    me,
		FromPlayer: me,
		Source:     ctx.Source(),
		Question:   t.ExileQuestion,
		Cards:      exiled,
		Min:        0,
		Max:        1,
		ThenKey:    t.Exile.then,
		CarryWord:  t.Word,
	})
	return nil
}
