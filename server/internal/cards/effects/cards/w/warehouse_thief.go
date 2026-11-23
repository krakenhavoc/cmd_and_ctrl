package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warehouse Thief — Creature — Tiefling Rogue {3}{R}, 4/2 (Commander Legends: Battle
// for Baldur's Gate):
//
//	"{2}, {T}, Sacrifice an artifact or creature: Exile the top card of your library.
//	 Until the end of your next turn, you may play that card."
//
// The cost is {2}, tap and a sacrifice of any artifact or creature, this creature
// included (nothing says "another"); the window is Prosper's "until the end of your next
// turn" (a seat-turn duration, so this turn's remainder counts too, CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddf34a38-29d5-4bf3-9b08-e7c9579a4200",
		Name:         "Warehouse Thief",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice an artifact or creature: Exile the top card of your library. Until the end of your next turn, you may play that card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{2}"), effects.TapCost(),
				effects.SacrificeN(1, "an artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))),
			Effect: effects.B20ExileTopUntilEndOfNextTurn,
		}},
	})
}
