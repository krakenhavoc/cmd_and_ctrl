package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haunting Voyage — Sorcery {4}{B}{B}:
//
//	"Choose a creature type. Return up to two creature cards of that type
//	 from your graveyard to the battlefield. If this spell was foretold,
//	 return all creature cards of that type from your graveyard to the
//	 battlefield instead.
//	 Foretell {5}{B}{B}"
//
// Whether the spell was foretold is read as it resolves and rides the tail
// as its N, since the choice is answered after the spell has left the stack.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "334c284e-7bc3-4c18-9f8d-4d846f80008e",
		Name:           "Haunting Voyage",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerGraveyard},
		SpecialActions: []game.SpecialAction{effects.Foretell("{5}{B}{B}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			then := ctx.ItemTail(hauntingVoyageTypeChosen)
			if ctx.SpellWasForetold() {
				then = effects.WithN(then, 1)
			}
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Haunting Voyage — choose a creature type", then)
			return nil
		},
	})
}

var hauntingVoyageTypeChosen = effects.CreatureTypeChosen("then/haunting-voyage-type-chosen", func(ctx *effects.Context, t string, o game.TailOutcome) error {
	match := func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(t) }
	if o.Tail.N > 0 {
		return effects.ReturnFromGraveyardTogether{
			Targets:    effects.GraveyardCardIDs(ctx, ctx.Controller(), match),
			Controller: ctx.Controller(),
		}.Apply(ctx)
	}
	return effects.ReturnChosenFromGraveyard{
		Question: "Haunting Voyage — return up to two creature cards of the chosen type to the battlefield",
		Match:    match,
		Min:      0,
		Max:      2,
	}.Apply(ctx)
})
