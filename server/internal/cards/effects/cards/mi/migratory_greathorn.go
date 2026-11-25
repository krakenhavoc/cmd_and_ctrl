package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Migratory Greathorn — Creature — Beast {3}{G}, 3/4 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {2}{G} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Whenever this creature mutates, search your library for a basic land card,
//	 put it onto the battlefield tapped, then shuffle."
//
// Mutate is the engine's (ADR 0223); the search is Cmdr2B15SearchBasicLand (a
// basic land card is read by supertype, and the player may fail to find, CR
// 701.23b; the shuffle follows either way).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4c6c8155-efa0-4c22-b255-34974a80c496",
		Name:             "Migratory Greathorn",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Lands: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Migratory Greathorn — search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				effects.Cmdr2B15SearchBasicLand("Migratory Greathorn — search for a basic land card", true)),
		},
	})
}
