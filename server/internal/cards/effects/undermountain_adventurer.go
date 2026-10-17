package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undermountain Adventurer — Creature — Giant Warrior {3}{G}, 3/4:
//
//	"Vigilance
//	 When this creature enters, you take the initiative.
//	 {T}: Add {G}{G}. If you've completed a dungeon, add six {G}
//	 instead."
//
// The ETB is TakeTheInitiative (ADR 0203, #4096). The mana ability's
// amount is read as it is activated (a ProducedFunc), from whether its
// controller has completed any dungeon this game (CR 309.7). A creature
// mana ability with {T} waits out summoning sickness (CR 302.6); the
// engine enforces that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2e5e0abd-38a6-46ca-b922-556db52f1332",
		Name:            "Undermountain Adventurer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Undermountain Adventurer"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: undermountainAdventurerProduced,
			Label:        "{T}: Add {G}{G}. If you've completed a dungeon, add six {G} instead.",
		}},
	})
}

// undermountainAdventurerProduced is {G}{G}, or six {G} for a player
// who has completed a dungeon.
func undermountainAdventurerProduced(g *game.Game, controller, _ uuid.UUID) string {
	if YouCompletedADungeon(g, controller) {
		return "{G}{G}{G}{G}{G}{G}"
	}
	return "{G}{G}"
}
