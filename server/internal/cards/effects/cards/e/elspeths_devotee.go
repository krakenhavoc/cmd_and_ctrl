package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth's Devotee — Creature — Human Soldier {2}{W}{W}, 3/3 (Theros Beyond
// Death, S82 OLD, #5514):
//
//	"When this creature enters, you may search your library and/or graveyard
//	 for a card named Elspeth, Undaunted Hero, reveal it, and put it into
//	 your hand. If you search your library this way, shuffle."
//
// Sorin's Guide's entry search (MayFetchNamedFromLibraryOrGraveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cacbdbcf-b5d2-40b3-bd8a-93636a248d3c",
		Name:         "Elspeth's Devotee",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.MayFetchNamedFromLibraryOrGraveyard("Elspeth's Devotee", "Elspeth, Undaunted Hero")},
	})
}
