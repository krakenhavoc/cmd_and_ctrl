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
// "Put into your graveyard" is a trip from the battlefield to the graveyard of
// Guan Yu's controller: its owner, when it is also its controller (a stolen
// one goes to its owner's graveyard, which is not "your" graveyard).
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
				effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.CardDied(ev, source) && source.Owner == source.Controller
				}, "Guan Yu — shuffle Guan Yu into your library", effects.ShuffleThisCardIntoItsOwnersLibrary),
				"Guan Yu — shuffle Guan Yu into your library?"),
		},
	})
}
