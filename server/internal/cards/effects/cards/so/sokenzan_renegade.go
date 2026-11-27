package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sokenzan Renegade — Creature — Ogre Samurai Mercenary {2}{R}, 3/3 (Saviors
// of Kamigawa, #5601):
//
//	"Bushido 1 (Whenever this creature blocks or becomes blocked, it gets
//	 +1/+1 until end of turn.)
//	 At the beginning of your upkeep, if a player has more cards in hand than
//	 each other player, the player who has the most cards in hand gains
//	 control of this creature."
//
// Wild Dogs' shape over hands: the leader is the one player whose hand is
// strictly larger than every other's still in the game; a tie at the top has
// no leader. The intervening "if" is checked as the upkeep begins and again as
// the ability resolves (CR 603.4), and control changes with no end
// (CR 611.2). If the leader is already its controller nothing changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63f414ed-01f4-428a-b65f-5c640679d5ef",
		Name:         "Sokenzan Renegade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Bushido(1, "Sokenzan Renegade"),
			effects.InterveningIf(effects.OldB592HandLeaderCondition(), effects.On(game.EventBeginUpkeep,
				effects.AllOf(effects.ByYou, effects.OldB592HandLeaderExists),
				"Sokenzan Renegade — the player with the most cards in hand gains control of this creature",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					leader, ok := effects.OldB592HandLeader(ctx.Game)
					if !ok {
						return nil
					}
					return effects.GainControl{
						Target:     ctx.Source(),
						Controller: leader,
						Duration:   game.IndefiniteDuration(),
						Label:      "Sokenzan Renegade — gain control",
					}.Apply(ctx)
				})))),
		},
	})
}
