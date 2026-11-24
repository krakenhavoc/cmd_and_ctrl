package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incarnation Technique — Sorcery {4}{B} (Secret Lair Drop):
//
//	"Demonstrate (When you cast this spell, you may copy it. If you do, choose
//	 an opponent to also copy it.)
//	 Mill five cards, then return a creature card from your graveyard to the
//	 battlefield."
//
// Demonstrate is the shared cast trigger (Demonstrate()). Each copy resolves
// for its own controller. Port of Karfell's order: the pick waits for the
// mill (MillToZone.ThenTail), so the cards just milled are on offer, and the
// return is a chosen card, not a target.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ab9539d-3461-4936-95cd-4c229308c4a4",
		Name:         "Incarnation Technique",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerGraveyard},
		Triggered:    []game.TriggeredAbility{effects.Demonstrate()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{Player: ctx.Controller(), N: 5, ThenTail: ctx.ItemTail(incarnationTechniqueMilled)}.Apply(ctx)
		},
	})
}

var incarnationTechniqueMilled = effects.CardTail("incarnation-technique/milled", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.ReturnChosenFromGraveyard{
		Player:   ctx.Controller(),
		Question: "Incarnation Technique — return a creature card from your graveyard to the battlefield",
		Match:    game.Card.IsCreature,
		Min:      1, Max: 1,
	}.Apply(ctx)
})
