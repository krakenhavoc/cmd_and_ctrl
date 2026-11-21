package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subterfuge — Creature — Elemental Incarnation {4}{U}, 3/5 (Lorwyn
// Eclipsed Commander):
//
//	"When this creature enters, target creature gains flying and "Whenever
//	 this creature deals combat damage to a player, draw that many cards"
//	 until end of turn.
//	 Encore {7}{U}{U} (...)"
//
// Flying is a keyword grant and the quoted trigger a duration grant (ADR
// 0093, Trash the Town's shape), both to the target as it resolves (CR
// 608.2b) and both lasting until end of turn. The trigger belongs to the
// creature, so its controller draws as many cards as the combat damage
// dealt. Encore is the shared keyword (encore.go).
//
// No simplification.
const subterfugeGrant = "subterfuge/combat-damage-draw-that-many"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "687cfdf7-8f35-464a-8d6e-9f2d769ed6b4",
		Name:         "Subterfuge",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: subterfugeGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Subterfuge — draw that many cards",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DrawCards{N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw that many cards.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Subterfuge — target creature gains flying and a draw trigger until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Subterfuge — flying"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantAbilitiesFor{
						Target: t.ID,
						Keys:   []string{subterfugeGrant},
						Label:  "Subterfuge — draws that many cards when it deals combat damage to a player",
					}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{effects.Encore("{7}{U}{U}")},
	})
}
