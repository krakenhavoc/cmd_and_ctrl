package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kenessos, Priest of Thassa — Legendary Creature — Merfolk Cleric {1}{U},
// 1/3 (Jumpstart 2022, #5948):
//
//	"If you would scry a number of cards, scry that many cards plus one
//	 instead.
//	 {3}{G/U}: Look at the top card of your library. If it's a Kraken,
//	 Leviathan, Octopus, or Serpent creature card, you may put it onto the
//	 battlefield. If you don't put the card onto the battlefield, you may
//	 put it on the bottom of your library."
//
// The first line is the scry replacement (CR 614.1, Eligeth-style ScryPlusOne),
// scoped to its controller. The look is private; the answer is one pick over
// what the rules allow (Bucolic Ranch's shape): put it onto the battlefield
// (only when it is a creature card of one of the four types, read as it sits in
// the library), put it on the bottom, or leave it on top (CR 608.2d). A token
// can't be in a library, so a card is always a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45b3a028-5705-4dc8-bfab-04bb5e01eea6",
		Name:         "Kenessos, Priest of Thassa",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.ScryPlusOne("Kenessos, Priest of Thassa: if you would scry a number of cards, scry that many cards plus one instead"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G/U}: Look at the top card of your library. If it's a Kraken, Leviathan, Octopus, or Serpent creature card, you may put it onto the battlefield. If you don't put the card onto the battlefield, you may put it on the bottom of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{G/U}"),
			Effect:  kenessosLook,
		}},
	})
}

const (
	kenessosPut    = "Put it onto the battlefield"
	kenessosBottom = "Put it on the bottom of your library"
	kenessosLeave  = "Leave it on top"
)

// kenessosLook looks at the top card and asks what to do with it.
func kenessosLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	top := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 1)
	if len(top) == 0 {
		return nil
	}
	options := []game.ChoiceOption{{Label: kenessosBottom}, {Label: kenessosLeave}}
	if c, ok := g.LookupCardForEffect(top[0]); ok && c.IsCreature() &&
		(c.HasSubtype(ctx.TextSubtype("Kraken")) || c.HasSubtype(ctx.TextSubtype("Leviathan")) ||
			c.HasSubtype(ctx.TextSubtype("Octopus")) || c.HasSubtype(ctx.TextSubtype("Serpent"))) {
		options = append([]game.ChoiceOption{{Label: kenessosPut}}, options...)
	}
	return effects.PickOption{
		Question: "Kenessos, Priest of Thassa — what do you do with the top card of your library?",
		Options:  options,
		ThenKey:  kenessosAnswer,
		Carry:    []uuid.UUID{top[0]},
	}.Apply(ctx)
}

// kenessosAnswer carries out the pick. An on-disk key: never renamed, never
// reused.
var kenessosAnswer = effects.OptionPickThen("option-pick/s82-gap-b154-kenessos-top-card", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	g, id := ctx.Game, r.Carry[0]
	switch r.Option.Label {
	case kenessosPut:
		_, err := g.PutFromLibraryOntoBattlefieldForEffect(id, game.LibraryEntryOptions{Controller: r.Chooser})
		return err
	case kenessosBottom:
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, []uuid.UUID{id})
	}
	return nil
})
