package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "de83b1a4-0ce0-4cc1-aff3-6d0239d92302",
		Name:         "Hostile Investigator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Hostile Investigator — target opponent discards a card", targetOpponentDiscards(1)),
				TargetPlayer("target opponent", Opponent())),
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventDiscardCard,
				func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventDiscardCard
				}, "Hostile Investigator — investigate", Do(Investigate{N: 1})))),
		},
	})
}
