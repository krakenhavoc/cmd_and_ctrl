package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "687cfdf7-8f35-464a-8d6e-9f2d769ed6b4",
		Name:         "Subterfuge",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: subterfugeGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Subterfuge — draw that many cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw that many cards.",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Subterfuge — target creature gains flying and a draw trigger until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Subterfuge — flying"}).Apply(ctx); err != nil {
						return err
					}
					return GrantAbilitiesFor{
						Target: t.ID,
						Keys:   []string{subterfugeGrant},
						Label:  "Subterfuge — draws that many cards when it deals combat damage to a player",
					}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{Encore("{7}{U}{U}")},
	})
}
