package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Svyelunite Priest — Creature — Merfolk Cleric {1}{U}, 1/1 (Fallen Empires, S82
// OLD batch 852, #5637):
//
//	"{U}{U}, {T}: Target creature gains shroud until end of turn. Activate only
//	 during your upkeep. (It can't be the target of spells or abilities.)"
//
// Abandon Reason's keyword grant on a target creature, behind Augur il-Vec's
// "only during your upkeep" window (the step and your own turn). The grant lands
// on the object the target is as the ability resolves; a target that has gone
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb7adbee-a3b0-42bf-988f-9d4faf6bd006",
		Name:         "Svyelunite Priest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{U}{U}, {T}: Target creature gains shroud until end of turn. Activate only during your upkeep.",
			Purpose:   effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "shroud")),
			Cost:      effects.Plus(effects.ManaCost("{U}{U}"), effects.TapCost()),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.DuringYourUpkeep(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"shroud"}, Label: "Svyelunite Priest — shroud"}.Apply(ctx)
			},
		}},
	})
}
