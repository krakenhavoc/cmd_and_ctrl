package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Command the Dreadhorde — Sorcery {4}{B}{B} (War of the Spark, #5520):
//
//	"Choose any number of target creature and/or planeswalker cards in
//	 graveyards. Command the Dreadhorde deals damage to you equal to the total
//	 mana value of those cards. Put them onto the battlefield under your
//	 control."
//
// The total is read from the cards that are still legal targets as the spell
// resolves (CR 608.2b), before they move, and the damage is dealt first, then
// every card enters together under the caster's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "198d31e4-1ae1-4bf9-a05a-7fa45c0aeddb",
		Name:         "Command the Dreadhorde",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		Targets: effects.TargetCardInGraveyard("any number of target creature and/or planeswalker cards in graveyards",
			effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			total := 0
			for _, id := range ids {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					if mv, ok := ctx.Game.ManaValueForEffect(c); ok {
						total += mv
					}
				}
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: total}).Apply(ctx); err != nil {
				return err
			}
			return effects.ReturnLegalGraveyardTargetsToBattlefieldTogether(ctx)
		},
	})
}
