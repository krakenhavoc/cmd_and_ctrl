package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Hunter — Creature — Human Warrior {W}, 2/1 (Dragons of Tarkir):
//
//	"Protection from Dragons
//	 This creature can block Dragons as though it had reach."
//
// The block waiver (#6250) bends only the flying check, for the Dragons the clause
// names: it blocks a flying Dragon and not a flying non-Dragon, and it gains no
// reach of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "11c8e218-0cd5-4356-89fd-f39f989e487b",
		Name:            "Dragon Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Dragons"},
		BlockRules:      []game.BlockRule{effects.CanBlockAsThoughItHadReach(effects.QuerySubtype("Dragon"))},
	})
}
