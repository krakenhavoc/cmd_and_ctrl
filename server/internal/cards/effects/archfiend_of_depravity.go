package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archfiend of Depravity — Creature — Demon {3}{B}{B}, 5/4 (EDHREC
// rank 1572):
//
//	"Flying
//	 At the beginning of each opponent's end step, that player chooses
//	 up to two creatures they control, then sacrifices the rest."
//
// The Demon that lets everyone keep two. The trigger fires on every
// opponent's end step (the event's Actor is the active player), and
// the choice is THAT PLAYER's: they are asked once, over their own
// creatures, for up to two to keep (keepCreaturesThenSacrificeTheRest),
// and then everything else they control is sacrificed at once. Nothing
// is targeted or destroyed, so hexproof and indestructible are
// irrelevant, as they are on paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "af247e2f-b271-4f5b-ab98-4579d2c17c21",
		Name:            "Archfiend of Depravity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b14OpponentsEndStepBegan(ev, source)
			},
			Key: "Archfiend of Depravity — that player keeps up to two creatures and sacrifices the rest",
			Effect: func(g *game.Game, item *game.StackItem) error {
				var victim uuid.UUID
				if item.Trigger != nil {
					victim = item.Trigger.Event.Actor
				}
				return keepCreaturesThenSacrificeTheRest(g, item, []uuid.UUID{victim},
					"Archfiend of Depravity — choose up to two creatures to keep; the rest are sacrificed", 2, nil)
			},
		}},
	})
}
