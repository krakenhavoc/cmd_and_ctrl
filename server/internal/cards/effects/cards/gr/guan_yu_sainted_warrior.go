package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guan Yu, Sainted Warrior — Legendary Creature — Human Soldier Warrior {3}{W}{W}, 3/5
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"When Guan Yu is put into your graveyard from the battlefield, you may shuffle Guan Yu into your library."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f7ee3935-4510-485a-a863-ac38dffa1908",
		Name:            "Guan Yu, Sainted Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.WhenThisDies("Guan Yu — shuffle Guan Yu into your library", effects.ShuffleThisCardIntoItsOwnersLibrary),
				"Guan Yu — shuffle Guan Yu into your library?"),
		},
	})
}
