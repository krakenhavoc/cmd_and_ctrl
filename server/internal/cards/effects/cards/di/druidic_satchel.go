package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Druidic Satchel — Artifact {3} (Commander 2021, #5504):
//
//	"{2}, {T}: Reveal the top card of your library. If it's a creature
//	 card, create a 1/1 green Saproling creature token. If it's a land
//	 card, put that card onto the battlefield under your control. If it's
//	 a noncreature, nonland card, you gain 2 life."
//
// The reveal is public and the card stays on top unless it is a land. The
// three clauses are independent "if"s over the one revealed card, so a
// creature land makes a Saproling AND enters the battlefield; a card that
// is neither a creature nor a land gains 2 life. An empty library reveals
// nothing and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75a4ddf5-1b3c-4dde-a6e4-4346e46cebf1",
		Name:         "Druidic Satchel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Reveal the top card of your library. If it's a creature card, create a 1/1 green Saproling creature token. If it's a land card, put that card onto the battlefield under your control. If it's a noncreature, nonland card, you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  druidicSatchelReveal,
		}},
	})
}

// druidicSatchelReveal is the ability's body.
func druidicSatchelReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you := ctx.Controller()
	revealed := g.RevealTopOfLibraryForEffect(you, ctx.Source(), 1, "Druidic Satchel — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	id := revealed[0]
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	isLand, isCreature := c.IsLand(), c.IsCreature()
	if isCreature {
		if err := (effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if !isCreature && !isLand {
		if err := (effects.GainLife{Player: you, Amount: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	if isLand {
		return effects.PutFromLibraryOntoBattlefield{
			Player: you,
			Cards:  revealed,
			Match:  func(_ *game.Game, _ uuid.UUID, card game.Card) bool { return card.InstanceID == id },
			All:    true,
		}.Apply(ctx)
	}
	return nil
}
