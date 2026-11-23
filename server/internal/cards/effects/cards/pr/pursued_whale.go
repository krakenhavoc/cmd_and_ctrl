package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pursued Whale — Creature — Whale {5}{U}{U}, 8/8 (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"When this creature enters, each opponent creates a 1/1 red Pirate
//	 creature token with "This token can't block" and "Creatures you control
//	 attack each combat if able."
//	 Spells your opponents cast that target this creature cost {3} more to
//	 cast."
//
// Each opponent, in turn order from the Whale's controller, creates one Pirate
// under their own control (the template in s82_gap_b222_tokens.go, whose
// second ability is Goblin Spymaster's). The tax is Monastery Siege's targeted
// cost increase with its predicate narrowed to the Whale itself: only SPELLS
// are taxed (the printed text names spells, not abilities), only opponents'
// spells, read off the targets announced.
//
// No simplification.
func init() {
	tax := effects.CostsMore(3,
		"Spells your opponents cast that target this creature cost {3} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	tax.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:      "6de2d53c-06e3-4c56-a5cf-c9cb2db42527",
		Name:          "Pursued Whale",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{tax},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pursued Whale — each opponent creates a 1/1 red Pirate creature token with \"This token can't block\" and \"Creatures you control attack each combat if able.\"",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range effects.OpponentsInTurnOrder(g, item.Controller) {
						if err := (effects.CreateToken{Controller: opp, Template: effects.S82GapB222PursuedPirateToken(), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
