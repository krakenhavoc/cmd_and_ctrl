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
// table waiting on it is a restore point. The exile pick before it still
// resumes through a closure: its "none" asks the outside pick with the
// card's word, which a keyed choose-cards pick has nowhere to carry.

// TakeFromExileOrOutside is the shape described above.
type TakeFromExileOrOutside struct {
	// Pick is the card's registered outside pick (registerOutsideTake,
	// revealing), whose filter is the kind of card ("an Eldrazi card",
	// "an artifact card") in exile too. Required.
	Pick outsideTake
	// Word is the word of the card's text the filter reads, if any.
	Word string
	// ExileQuestion heads the first prompt; OutsideQuestion the second.
	ExileQuestion, OutsideQuestion string
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
	outside := TakeFromOutside{Player: me, Pick: t.Pick, Word: t.Word, Question: t.OutsideQuestion}
	if len(exiled) == 0 {
		return outside.Apply(ctx)
	}
	item := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    me,
		FromPlayer: me,
		Source:     ctx.Source(),
		Question:   t.ExileQuestion,
		Cards:      exiled,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneExile,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 1 {
				return nothingIfGone(g.BounceToHandForEffect(picked[0]))
			}
			return outside.Apply(NewContext(g, item))
		},
	})
	return nil
}
