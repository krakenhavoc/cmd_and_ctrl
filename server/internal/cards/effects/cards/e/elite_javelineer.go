package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Javelineer — Creature — Human Soldier {2}{W}, 2/2 (Eighth Edition, S82
// OLD batch 610, #5607):
//
//	"Whenever this creature blocks, it deals 1 damage to target attacking
//	 creature."
//
// One trigger per block (a creature that blocks several attackers blocks once,
// CR 509.3a). The target is chosen as the trigger goes on the stack (CR 603.3d)
// and must be an attacking creature then and as it resolves (CR 608.2b); the
// Javelineer deals the damage as the object it was, so a Javelineer that has
// left the battlefield by then still deals it (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "649300e2-c491-435d-92d5-f1b88f0e2e2e",
		Name:         "Elite Javelineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Elite Javelineer — it deals 1 damage to target attacking creature", effects.OldB610ThisDealsOneToTheTarget),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())),
		},
	})
}
