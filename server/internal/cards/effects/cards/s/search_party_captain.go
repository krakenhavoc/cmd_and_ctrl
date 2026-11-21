package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Search Party Captain — Creature — Human Soldier {3}{W}, 2/2 (Midnight
// Hunt, #5497):
//
//	"This spell costs {1} less to cast for each creature you attacked with
//	 this turn.
//	 When this creature enters, draw a card."
//
// The discount is read at CR 601.2f off this turn's attack record
// (Game.DistinctAttackersThisTurn): a creature that attacked in two combat
// phases counts once, and a creature put onto the battlefield attacking was
// never declared as an attacker (CR 508.4), so it does not count. It spends
// generic mana only, so the Captain never costs less than {W}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8721b0a6-52ff-4749-92db-82bb36f5a73e",
		Name:         "Search Party Captain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return q.Game.DistinctAttackersThisTurn(q.Controller)
			}, "This spell costs {1} less to cast for each creature you attacked with this turn."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Self, "Search Party Captain — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
