package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "76c97e3d-a9be-4ed7-a2ef-f97870d8f58a",
		Name:            "Feldon, Ronom Excavator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Feldon — exile that many cards from the top of your library, choose one, you may play it until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.S82Bro13ExileChooseOnePlay(effects.NewContext(g, item), item.Trigger.Event.Amount)
				}),
		},
	})
}
