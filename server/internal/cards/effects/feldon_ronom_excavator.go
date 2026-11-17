package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feldon, Ronom Excavator — Legendary Creature — Human Artificer {1}{R}, 2/2
// (The Brothers' War, #5482):
//
//	"Haste
//	 Feldon can't block.
//	 Whenever Feldon is dealt damage, exile that many cards from the top of
//	 your library. Choose one of them. Until the end of your next turn, you
//	 may play that card."
//
// Boros Reckoner's trigger: it waits for the damage to finish and reads the
// TOTAL Feldon was dealt (CR 120.3), even if he died to it. The exile, the
// choice and the window are s82Bro13ExileChooseOnePlay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "76c97e3d-a9be-4ed7-a2ef-f97870d8f58a",
		Name:            "Feldon, Ronom Excavator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Feldon — exile that many cards from the top of your library, choose one, you may play it until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return s82Bro13ExileChooseOnePlay(NewContext(g, item), item.Trigger.Event.Amount)
				}),
		},
	})
}
