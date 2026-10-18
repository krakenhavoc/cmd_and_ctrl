package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imoen, Mystic Trickster — Legendary Creature — Human Rogue Wizard
// {2}{U}, 2/3:
//
//	"Ward {2}
//	 At the beginning of your end step, if you have the initiative,
//	 draw a card. Draw another card if you've completed a dungeon.
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// Ward is the per-card Ward helper. The end-step clause is an
// intervening "if" on the initiative (CR 603.4, IfYouHaveTheInitiative,
// ADR 0203 / #4096); "draw another card if you've completed a dungeon"
// is a clause of the effect, read as it resolves (CR 608.2), and counts
// any dungeon completed this game (CR 309.7). Choose a Background is a
// deck-building rule the deck validator applies (deck/partner.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "280e05c5-7ab4-45bf-aa15-e594cec49ef3",
		Name:         "Imoen, Mystic Trickster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Imoen, Mystic Trickster — ward {2}"),
			IfYouHaveTheInitiative(AtYourEndStep("Imoen, Mystic Trickster — draw a card, and another if you've completed a dungeon",
				imoenDraw)),
		},
	})
}

// imoenDraw draws one card, or two for a player who has completed a
// dungeon.
func imoenDraw(g *game.Game, item *game.StackItem) error {
	n := 1
	if YouCompletedADungeon(g, item.Controller) {
		n = 2
	}
	return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
}
