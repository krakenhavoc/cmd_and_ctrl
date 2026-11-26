package cardsdi

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Draco — Artifact Creature — Dragon {16}, 9/9 (From the Vault: Dragons, S82 OLD
// batch 519, #5589):
//
//	"Domain — This spell costs {2} less to cast for each basic land type among
//	 lands you control.
//	 Flying
//	 Domain — At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {10}. This cost is reduced by {2} for each basic land type among lands you
//	 control."
//
// The cast reduction is Yavimaya Sojourner's DomainCostReduction at two per
// type (generic mana only, CR 601.2f; five types leave {6}). The upkeep payment
// is Wild Leotau's pay-unless with a cost read as the trigger resolves: {10}
// less {2} for each basic land type its controller controls then, never below
// {0} (CR 107.1b); a cost of {0} is paid with nothing asked. Declining, or being
// unable to pay, sacrifices it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6157ade6-9037-4339-869b-493fdee71dcf",
		Name:            "Draco",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return 2 * effects.DomainCostAmount(q) },
				"Domain — This spell costs {2} less to cast for each basic land type among lands you control"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Draco — sacrifice unless you pay {10}, reduced by {2} for each basic land type among lands you control",
				func(g *game.Game, item *game.StackItem) error {
					n := max(0, 10-2*effects.DomainCount(g, item.Controller))
					if n == 0 {
						return nil
					}
					cost := fmt.Sprintf("{%d}", n)
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     cost,
						Question: "Draco — pay " + cost + " or sacrifice Draco?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
