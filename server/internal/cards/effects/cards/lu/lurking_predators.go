package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lurking Predators — Enchantment {4}{G}{G}:
//
//	"Whenever an opponent casts a spell, reveal the top card of your
//	 library. If it's a creature card, put it onto the battlefield.
//	 Otherwise, you may put that card on the bottom of your library."
//
// Every opposing spell is a free creature or a free look-and-bin. The
// reveal and the put are #745's library-to-battlefield move — no
// search, no shuffle, a real entry through the CR 614 pipeline. The
// "otherwise" is a yes/no to the enchantment's controller; yes moves
// the card to the bottom, no leaves it on top where the whole table
// has now seen it.
//
// A creature a replacement keeps off the battlefield is not offered the
// bottom: "otherwise" is about what the card is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15fbb7b1-c62d-4f82-9f35-2c10299779f4",
		Name:         "Lurking Predators",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(nil),
				"Lurking Predators — reveal the top card; a creature goes onto the battlefield",
				lurkingPredatorsReveal),
		},
	})
}

func lurkingPredatorsReveal(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	top, err := effects.RevealTopThenPutIfMatch(g, source, controller, game.Card.IsCreature, false,
		"Lurking Predators — revealed from the top of the library")
	if err != nil || top == uuid.Nil {
		return err
	}
	c, ok := g.LookupCardForEffect(top)
	if !ok || (c.IsCreature() && !effects.IsToken(c)) {
		return nil
	}
	return effects.MayChoice{
		Player:   controller,
		Question: "Lurking Predators — put " + c.Name + " on the bottom of your library?",
		YesLabel: "Put it on the bottom",
		NoLabel:  "Leave it on top",
		Then:     LurkingPredatorsBottom,
		Carry:    []uuid.UUID{top},
	}.Apply(effects.NewContext(g, item))
}

// lurkingPredatorsBottom is "you may put that card on the bottom of your
// library" (#4298: keyed; the card is the prompt's carried ID).
var LurkingPredatorsBottom = effects.MayYes("confirm/lurking-predators-bottom", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	// A pile of one: the random-order bottom is the plain bottom, and it is
	// the move that repositions a card already in its owner's library.
	// "That card" is the one on top of the library: ZoneLibrary makes a card
	// that left it before the answer arrived (drawn, milled) a no-op rather
	// than pulling it back out of a hand or a graveyard.
	return ctx.Game.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, []uuid.UUID{effects.CarryAt(r, 0)})
})
