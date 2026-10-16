package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside_game.go — "a card you own from outside the game" (CR 400.11,
// ADR 0191). A deck's sideboard is that pool; this file is the effect
// side of it. The only way a card leaves the sideboard is an effect
// that brings one in (CR 400.11c).

// TakeFromOutside is "You may reveal a <kind> card you own from outside
// the game and put it into your hand" (the wishes) and its unrevealed,
// mandatory sibling "Put a card you own from outside the game into your
// hand" (Mastermind's Acquisition).
//
// It asks the owner to choose from their sideboard (a choose_cards
// prompt over cards only they can see) and brings the choice into their
// hand. A player whose deck seated no sideboard chooses from the card
// database instead (ADR 0198): the same filter, minted into a real card
// they own, and the same reveal and move. With nothing to choose from it does nothing and asks nothing: a
// clause that finds no card is not an error (CR 608.2c's "as much as
// possible"). While another player controls the owner's turn the pool is
// closed (ADR 0167 §5, CR 723.4) and it finds nothing.
type TakeFromOutside struct {
	// Player owns the cards and chooses. Zero means the resolving
	// effect's controller.
	Player uuid.UUID
	// Match narrows the pool ("a sorcery card"); nil offers every card.
	// It runs when the prompt is queued, never later.
	Match func(game.Card) bool
	// Question is the prompt's header, written the way the card is.
	Question string
	// Required is "put a card" rather than "you may reveal a card": the
	// owner may not decline when something matches.
	Required bool
	// Reveal shows the chosen card to the table before it moves. Every
	// printed wish reveals; Mastermind's Acquisition does not.
	Reveal bool
}

func (t TakeFromOutside) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source := ctx.Source()
	question, reveal := t.Question, t.Reveal
	floor := 0
	if t.Required {
		floor = 1
	}
	// Frozen scalars only (the StackItem.Effect contract): the closure
	// resolves against whichever *Game it is handed.
	ctx.Game.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:   player,
		Source:   source,
		Question: question,
		Match:    t.Match,
		Min:      floor,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			if reveal {
				g.RevealOutsideForEffect(game.RevealSpec{
					Player: player,
					Source: source,
					Reason: question,
					Cards:  picked,
				})
			}
			return g.BringFromOutsideForEffect(player, picked[0])
		},
	})
	return nil
}
