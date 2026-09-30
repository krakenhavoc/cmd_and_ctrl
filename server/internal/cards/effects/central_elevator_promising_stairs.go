package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Central Elevator // Promising Stairs — Enchantment — Room (ADR 0103):
//
//	Central Elevator {3}{U}: "When you unlock this door, search your
//	  library for a Room card that doesn't have the same name as a Room
//	  you control, reveal it, put it into your hand, then shuffle."
//	Promising Stairs {2}{U}: "At the beginning of your upkeep, surveil
//	  1. You win the game if there are eight or more different names
//	  among unlocked doors of Rooms you control."
//
// "The same name" is CR 201.2's: two objects share a name when they
// have one name in common, so a library Room is excluded when EITHER of
// its names is the name of an unlocked door of a Room you control
// (UnlockedDoorNamesYouControl; a Room's locked half has no name,
// CR 709.5). The set is read when the search resolves.
//
// The win is checked after the surveil, in its continuation ("surveil
// 1. You win ..." reads the doors once the surveil is done), through
// WinTheGame, so a Platinum Angel stops it (ADR 0057).
//
// No simplification.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "927fa223-69a3-4563-80b1-b578f3d031f1",
		Name:         "Central Elevator // Promising Stairs",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Central Elevator — search for a Room card with a new name", centralElevatorSearch),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Promising Stairs — surveil 1, then win with eight different door names", promisingStairsUpkeep),
		}},
	}))
}

func centralElevatorSearch(g *game.Game, item *game.StackItem) error {
	taken := map[string]bool{}
	for _, n := range UnlockedDoorNamesYouControl(g, item.Controller) {
		taken[n] = true
	}
	return SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			if !c.HasSubtype("Room") {
				return false
			}
			for _, n := range game.NamesOf(c) {
				if taken[n] {
					return false
				}
			}
			return true
		},
		Dest:    game.ZoneHand,
		Limit:   1,
		Reveal:  true,
		Shuffle: true,
		Reason:  "Central Elevator — a Room card that doesn't share a name with a Room you control",
	}.Apply(NewContext(g, item))
}

func promisingStairsUpkeep(g *game.Game, item *game.StackItem) error {
	me := item.Controller
	return Surveil{Player: me, N: 1, Then: func(g *game.Game) error {
		if distinctDoorNames(g, me) < 8 {
			return nil
		}
		return WinTheGame{Player: me}.Apply(NewContext(g, item))
	}}.Apply(NewContext(g, item))
}
