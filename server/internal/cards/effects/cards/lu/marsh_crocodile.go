package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marsh Crocodile — Creature — Crocodile {2}{U}{B}, 4/4 (Planeshift, S82 OLD b643, #5615):
//
//	"When this creature enters, return a blue or black creature you control to its
//	 owner's hand.
//	 When this creature enters, each player discards a card."
//
// Two printed triggers, two registered triggers. The return is Whitemane Lion's
// mandatory choice on resolution (not a target), and the Crocodile itself is a
// legal pick while it is on the battlefield (it is blue and black). Each player
// then chooses their own discard (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d6f3934-d8f1-4758-b87a-8df8d7dc492c",
		Name:         "Marsh Crocodile",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Marsh Crocodile — return a blue or black creature you control to its owner's hand", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.OldB643CreatureOfEitherColor("U", "B"),
				Question: "Marsh Crocodile — return a blue or black creature you control to its owner's hand",
			})),
			effects.WhenThisEnters("Marsh Crocodile — each player discards a card", effects.OldB643EachPlayerDiscards("Marsh Crocodile")),
		},
	})
}
