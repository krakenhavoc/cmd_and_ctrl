package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sin Collector — Creature — Human Cleric {1}{W}{B}, 2/1 (Dragon's Maze, S82
// OLD b406, #5564):
//
//	"When this creature enters, target opponent reveals their hand. You choose
//	 an instant or sorcery card from it and exile that card."
//
// Thought-Knot Seer's enters trigger (TargetRevealsYouChooseExile) with the
// pick narrowed to instant and sorcery cards. The exile is permanent; a hand
// with no instant or sorcery is revealed and nothing is exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bef2904-1ebe-4e7e-879a-79f4968174fb",
		Name:         "Sin Collector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sin Collector — target opponent reveals their hand; you choose an instant or sorcery card from it and exile it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.TargetRevealsYouChooseExile(effects.Or(effects.Instant(), effects.Sorcery()), "instant or sorcery card")(item, effects.NewContext(g, item))
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
