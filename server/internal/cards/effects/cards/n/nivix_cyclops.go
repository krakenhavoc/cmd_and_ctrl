package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nivix Cyclops — Creature — Cyclops {1}{U}{R}, 1/4 (Dragon's Maze, S82 OLD batch
// 404, #5564):
//
//	"Defender
//	 Whenever you cast an instant or sorcery spell, this creature gets +3/+0 until
//	 end of turn and can attack this turn as though it didn't have defender."
//
// One trigger on each instant or sorcery you cast (CR 603.2c), Erratic Cyclops's
// shape: a layer-7 +3/+0 until end of turn and Hightide Hermit's defender waiver
// (CR 702.3b: it may attack this turn as though it didn't have defender), both on
// this creature only if it is still the same object on the battlefield.
//
// No simplification.
func init() {
	canAttack := effects.CanAttackThisTurnAsThoughNoDefender("Nivix Cyclops — can attack this turn as though it didn't have defender")
	effects.Register(effects.Spec{
		OracleID:        "68b8bfac-63f7-4dbd-bb2e-4ad394c77335",
		Name:            "Nivix Cyclops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Nivix Cyclops — +3/+0 until end of turn and can attack as though it didn't have defender",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.ThisGetsUntilEndOfTurn(3, 0, "Nivix Cyclops — +3/+0 until end of turn")(g, item); err != nil {
						return err
					}
					return canAttack(g, item)
				}),
		},
	})
}
