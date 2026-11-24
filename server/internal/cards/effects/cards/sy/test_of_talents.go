package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Test of Talents — Instant {1}{U} (Strixhaven: School of Mages, #5502):
//
//	"Counter target instant or sorcery spell. Search its controller's
//	 graveyard, hand, and library for any number of cards with the same name
//	 as that spell and exile them. That player shuffles, then draws a card for
//	 each card exiled from their hand this way."
//
// Quash's body with The End's tail. The spell's name and controller are read
// before it is countered, while it is on the stack (a split card cast as one
// half has only that half's name, CR 709.3b); the countered card then lies in
// the graveyard where the search finds it. "Any number" lets the caster leave
// matches in the hidden zones. A spell that can't be countered stays on the
// stack and the search still happens; a target that has left the stack makes
// the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d808262d-43ca-46f8-ab8b-b584773f9c58",
		Name:         "Test of Talents",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target instant or sorcery spell", effects.Or(effects.Instant(), effects.Sorcery())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			spell := ctx.Game.StackItemForEffect(target)
			c, found := ctx.Game.LookupCardForEffect(target)
			if spell == nil || !found {
				return nil
			}
			victim, names := spell.Controller, effects.ExtractionNames(c)
			if err := (effects.CounterTarget{StackID: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.OldB43ExtractAndVictimDraws(ctx, "Test of Talents", victim, names)
		},
	})
}
