package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "334c284e-7bc3-4c18-9f8d-4d846f80008e",
		Name:           "Haunting Voyage",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerGraveyard},
		SpecialActions: []game.SpecialAction{Foretell("{5}{B}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			then := ctx.ItemTail(hauntingVoyageTypeChosen)
			if ctx.SpellWasForetold() {
				then = withN(then, 1)
			}
			ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Haunting Voyage — choose a creature type", then)
			return nil
		},
	})
}

var hauntingVoyageTypeChosen = creatureTypeChosen("then/haunting-voyage-type-chosen", func(ctx *Context, t string, o game.TailOutcome) error {
	match := func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(t) }
	if o.Tail.N > 0 {
		return ReturnFromGraveyardTogether{
			Targets:    graveyardCardIDs(ctx, ctx.Controller(), match),
			Controller: ctx.Controller(),
		}.Apply(ctx)
	}
	return ReturnChosenFromGraveyard{
		Question: "Haunting Voyage — return up to two creature cards of the chosen type to the battlefield",
		Match:    match,
		Min:      0,
		Max:      2,
	}.Apply(ctx)
})
