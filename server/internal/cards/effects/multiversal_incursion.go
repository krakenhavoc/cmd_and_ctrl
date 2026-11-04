package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Multiversal Incursion — Sorcery {5}{U}{U} (Marvel Super Heroes, #4455):
//
//	"For each nontoken creature you control, create a token that's a copy
//	 of that creature, except it isn't legendary."
//
// Kindred Charge's shape: the originals are listed once, when the spell
// resolves, so a token this makes is never copied in turn, and a token you
// already control is never one of them. Each copy is Irenicus's Vile
// Duplication's: a token copy of the object (copiable values, the original's
// triggers and abilities along) with the legendary supertype taken off its
// type line, part of the copy's copiable values (CR 707.9b). A creature that
// has left by then is skipped. Every copy is made by ONE creation, one group
// per original (#4594, CreateTokenCopiesTogether): the tokens are created
// at the same time (CR 608.2f) and enter together, so each newcomer sees
// the others enter (CR 603.6a) and a "one or more" trigger counts one
// event (CR 603.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3678fce-6897-4f2c-bf50-243facb6581b",
		Name:         "Multiversal Incursion",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var originals []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsCreature() && !c.IsToken() {
					originals = append(originals, c.InstanceID)
				}
			}
			return CreateTokenCopiesTogether{
				Controller: item.Controller,
				Copies:     originals,
				Except: func(tok *game.Card) {
					tok.TypeLine = removeLegendaryFromTypeLine(tok.TypeLine)
				},
			}.Apply(ctx)
		},
	})
}
