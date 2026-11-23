package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Counterpoint — Instant {3}{U}{B}:
//
//	"Counter target spell. You may cast a creature, instant, sorcery, or
//	 planeswalker spell from your graveyard with mana value less than or
//	 equal to that spell's mana value without paying its mana cost."
//
// The countered spell's mana value is read while it is still on the stack
// (CR 202.3e, so an X spell counts its X), before it is countered. The cast
// is made during the resolution (CR 608.2g) through the resolution_cast
// prompt over your graveyard's qualifying cards, one cast. The countered
// spell, now in its owner's graveyard, is itself a candidate when it is
// yours. The cap is judged on the face being cast.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dad4af15-5b26-417d-a41b-29e671e878f4",
		Name:         "Counterpoint",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			mv, known := 0, false
			if ctx.Game.StackItemForEffect(stackID) != nil {
				if card, ok := ctx.Game.LookupCardForEffect(stackID); ok {
					mv, _ = ctx.Game.ManaValueForEffect(card)
					known = true
				}
			}
			if err := (effects.CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if !known {
				return nil
			}
			spell := effects.FreeCastSpell{Filter: game.PermissionFilter{CardTypesAny: "creature instant sorcery planeswalker"}}.
				WithManaValueAtMost(mv)
			clause := spell.On(game.CastPermission{})
			effects.Cmdr2S42FreeGraveyardCast(ctx, "Counterpoint", spell, false,
				effects.GraveyardCardIDs(ctx, ctx.Controller(), clause.OffersASpell))
			return nil
		},
	})
}
