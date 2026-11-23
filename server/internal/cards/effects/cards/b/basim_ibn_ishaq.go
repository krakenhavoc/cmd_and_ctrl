package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basim Ibn Ishaq — Legendary Creature — Human Assassin {U}{B}, 2/2
// (Assassin's Creed, #5948):
//
//	"Whenever you cast a historic spell, draw a card. Basim Ibn Ishaq can't
//	 be blocked this turn. This ability triggers only once each turn.
//	 (Artifacts, legendaries, and Sagas are historic.)
//	 Whenever Basim Ibn Ishaq deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// The first ability is Jhoira's historic cast watch under Academy Wall's
// once-each-turn guard (spent as the trigger is created, so a countered
// ability still used the turn), drawing a card and then making Basim unblockable
// until end of turn (CR 611.2c: the restriction is applied as it resolves).
// The second is a counter on Basim himself when he connects; Basim gone by
// then gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a70bde5-c514-4dcf-919a-48bdd1c17768",
		Name:         "Basim Ibn Ishaq",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.On(game.EventCast, effects.YouCastAHistoricSpell,
				"Basim Ibn Ishaq — draw a card; Basim can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Basim Ibn Ishaq can't be blocked this turn"}.Apply(ctx)
				})),
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Basim Ibn Ishaq — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
