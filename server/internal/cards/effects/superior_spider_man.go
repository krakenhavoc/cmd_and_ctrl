package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Superior Spider-Man — Legendary Creature — Spider Human Hero {2}{U}{B},
// 4/4 (Marvel's Spider-Man, pspm, #4460):
//
//	"Mind Swap — You may have Superior Spider-Man enter as a copy of any
//	 creature card in a graveyard, except his name is Superior Spider-Man
//	 and he's a 4/4 Spider Human Hero in addition to his other types. When
//	 you do, exile that card."
//
// An enters-as-a-copy of a graveyard card (#4797, CR 707.2, CR 614.1c).
// The except clause is part of the copy effect (CR 707.9b): the name, a
// base 4/4 and the three creature types. Whether the copy is legendary is
// the copied card's (the clause adds none).
//
// "When you do, exile that card" is a reflexive trigger the replacement
// creates (CR 603.12). The except clause declares it on the entry event
// with the copied card's object, and the engine creates it once the
// permanent has landed, so it goes on the stack after the entry and a
// player can answer it (the copy is already settled, so exiling the card
// changes nothing about it). It exiles the card only if it is still that
// object in a graveyard (CR 400.7). Declined, Superior Spider-Man enters
// as his printed 4/4 self and nothing triggers.
func init() {
	Register(Spec{
		OracleID:     "636cc915-9d1b-4ffe-9e74-795b78663911",
		Name:         "Superior Spider-Man",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Superior Spider-Man", anyCreatureCardInAGraveyard, superiorSpiderManExcept),
		},
	})
}

// superiorSpiderManExcept is the except clause and the "when you do".
func superiorSpiderManExcept(ev *game.ReplacementEvent, v *game.PrintedValues, g *game.Game, source *game.Card) {
	v.SetName("Superior Spider-Man")
	v.Power, v.Toughness = 4, 4
	v.AddSubtype("Spider")
	v.AddSubtype("Human")
	v.AddSubtype("Hero")
	copied := ev.CopySourceForEffect()
	c, ok := g.LookupCardForEffect(copied)
	if !ok {
		return
	}
	// The declaration is refused only for a malformed one, which this is
	// not; there is nothing for a copy to do about a refusal anyway.
	_ = g.QueueReflexiveTriggerAfterReplacedMoveForEffect(ev, source, game.ReflexiveTrigger{
		Label:  "Superior Spider-Man — exile that card",
		Body:   superiorSpiderManExileBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: copied, Epoch: c.ObjectEpoch}},
	})
}

var superiorSpiderManExileBody = game.ReflexiveBody("superior-spider-man/exile-that-card", simpleBody(func(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	if !g.CardInGraveyardForEffect(ref) {
		return nil
	}
	if err := g.ExileCardForEffect(ref.ID); err != nil && !errors.Is(err, game.ErrCardNotFound) {
		return err
	}
	return nil
}), nil)
