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

// TakeFromExileOrOutside is the shape described above.
type TakeFromExileOrOutside struct {
	// Match is the kind of card ("an Eldrazi card", "an artifact card").
	Match func(game.Card) bool
	// ExileQuestion heads the first prompt; OutsideQuestion the second.
	ExileQuestion, OutsideQuestion string
}

func (t TakeFromExileOrOutside) Apply(ctx *Context) error {
	me := ctx.Controller()
	var exiled []uuid.UUID
	if ex := ctx.Game.Exile; ex != nil {
		for _, c := range ex.Cards {
			if c.Owner == me && !c.FaceDown && (t.Match == nil || t.Match(c)) {
				exiled = append(exiled, c.InstanceID)
			}
		}
	}
	outside := TakeFromOutside{Player: me, Match: t.Match, Question: t.OutsideQuestion, Reveal: true}
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
