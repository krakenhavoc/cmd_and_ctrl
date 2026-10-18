package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Safana, Calimport Cutthroat — Legendary Creature — Human Rogue
// {2}{B}, 3/2:
//
//	"Menace
//	 At the beginning of your end step, if you have the initiative,
//	 create a Treasure token. Create three of those tokens instead if
//	 you've completed a dungeon.
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// The end-step clause is an intervening "if" on the initiative (CR
// 603.4, IfYouHaveTheInitiative, ADR 0203 / #4096); "three instead if
// you've completed a dungeon" is a clause of the effect, read as it
// resolves (CR 608.2). Choose a Background is a deck-building rule the
// deck validator applies (deck/partner.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "afc1c443-93e3-4ef4-a404-1d9fc13e0b02",
		Name:            "Safana, Calimport Cutthroat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			IfYouHaveTheInitiative(AtYourEndStep("Safana, Calimport Cutthroat — create a Treasure, or three if you've completed a dungeon",
				safanaTreasure)),
		},
	})
}

// safanaTreasure creates one Treasure, or three for a player who has
// completed a dungeon.
func safanaTreasure(g *game.Game, item *game.StackItem) error {
	n := 1
	if YouCompletedADungeon(g, item.Controller) {
		n = 3
	}
	return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: n}.Apply(NewContext(g, item))
}
