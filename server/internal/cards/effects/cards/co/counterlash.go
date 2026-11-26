package cardsco

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Counterlash — Instant {4}{U}{U}:
//
//	"Counter target spell. You may cast a spell that shares a card type with
//	 it from your hand without paying its mana cost."
//
// The card types are read off the spell before it is countered (it is gone
// from the stack afterwards). The free cast is judged on the face being cast
// (CardTypesAny: a card qualifies with at least one of the types, CR 205.2a).
func init() {
	offer := effects.RegisterFreeHandCastOffer("choose-cards/counterlash-cast", "Counterlash")
	effects.Register(effects.Spec{
		OracleID:     "16de96d0-a58e-42a7-9a13-45358542a246",
		Name:         "Counterlash",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			var shared []string
			if ctx.Game.StackItemForEffect(stackID) != nil {
				if spell, ok := ctx.Game.LookupCardForEffect(stackID); ok {
					for _, t := range effects.OldB443CardTypes {
						if spell.HasCardType(t) {
							shared = append(shared, t)
						}
					}
				}
			}
			if err := (effects.CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if len(shared) == 0 {
				return nil
			}
			effects.OfferFreeHandCast(ctx, ctx.Controller(), offer, "a spell that shares a card type with it",
				effects.FreeCastSpell{Filter: game.PermissionFilter{CardTypesAny: strings.Join(shared, " ")}})
			return nil
		},
	})
}
