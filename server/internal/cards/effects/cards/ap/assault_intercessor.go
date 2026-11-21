package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assault Intercessor — Creature — Astartes Warrior {1}{W}{B}, 3/2:
//
//	"First strike, menace
//	 Chainsword — Whenever a creature an opponent controls dies, that
//	 player loses 2 life."
//
// Archfiend of the Dross's row: the creature's controller as it last
// existed on the battlefield (CR 603.10a) loses the life, and it is life
// loss, not damage.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ba19659-461f-482c-b882-b4fca7d5d1f8",
		Name:            "Assault Intercessor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "menace"},
		Triggered:       []game.TriggeredAbility{effects.S82GapB191OpponentCreatureDiesTrigger("Assault Intercessor — that player loses 2 life", 2)},
	})
}
