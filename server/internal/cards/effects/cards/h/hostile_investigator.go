package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hostile Investigator — Creature — Ogre Rogue Detective {3}{B}, 4/4 (The
// Big Score, #4468):
//
//	"When this creature enters, target opponent discards a card.
//	 Whenever one or more players discard one or more cards, investigate.
//	 This ability triggers only once each turn."
//
// The second ability reads every discard by any player (including its own
// enters trigger's). A batch of discards (a hand dumped at once) is one
// occurrence (OncePerBatch), and the once-a-turn limit is the engine's
// per-object tally (TriggersOnlyOnceEachTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de83b1a4-0ce0-4cc1-aff3-6d0239d92302",
		Name:         "Hostile Investigator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Hostile Investigator — target opponent discards a card", effects.TargetOpponentDiscards(1)),
				effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventDiscardCard,
				func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventDiscardCard
				}, "Hostile Investigator — investigate", effects.Do(effects.Investigate{N: 1})))),
		},
	})
}
