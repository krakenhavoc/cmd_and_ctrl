package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromoka Warrior — Creature — Human Warrior {1}{W}, 3/1 (Dragons of Tarkir,
// #5551):
//
//	"(no rules text)"
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f62998a7-898f-494f-89a1-624f6236033b",
		Name:         "Dromoka Warrior",
		Completeness: effects.CompletenessFull,
	})
}
