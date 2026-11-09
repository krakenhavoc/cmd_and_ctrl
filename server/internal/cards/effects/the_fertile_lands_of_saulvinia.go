package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fertile Lands of Saulvinia — Plane — Antausia (Commander Masters,
// #4870):
//
//	"Whenever a player taps a land for mana, that player adds one mana of
//	 any type that land produced.
//	 Whenever chaos ensues, reveal cards from the top of your planar deck
//	 until you reveal a plane card. Chaos ensues on that plane. Then put all
//	 cards revealed this way on the bottom of your planar deck in any
//	 order."
//
// A plane (ADR 0216 §6). The first ability is a triggered mana ability (CR
// 605.1b) that functions from the command zone (CR 901.7) and gives the
// player who tapped the land one more mana of any type it produced. The chaos
// ability reveals from the planar deck until a plane card (phenomena are
// revealed past), and chaos ensues on that plane only (CR 311.7): its chaos
// ability triggers while it is still on the planar deck, revealed, and this
// plane's does not (the ruling). Then every revealed card, that plane
// included, goes to the bottom of the planar deck in an order the planar
// controller chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f3e71b6-0fd6-4ac5-a265-faaabe21177b",
		Name:         "The Fertile Lands of Saulvinia",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			WheneverAPlayerTapsALandForMana(
				"The Fertile Lands of Saulvinia — add one mana of any type that land produced",
				AddsOneManaOfAnyTypeProduced()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("The Fertile Lands of Saulvinia — reveal until a plane card; chaos ensues on that plane; put the revealed cards on the bottom of your planar deck in any order",
				func(g *game.Game, item *game.StackItem) error {
					run, plane := g.RevealPlanarUntilPlaneForEffect(item.Controller, item.SourceCardID, "The Fertile Lands of Saulvinia — revealed until a plane card")
					if plane != uuid.Nil {
						g.ChaosEnsuesForEffect(item.Controller, plane)
					}
					g.PutPlanarCardsOnBottomInAnyOrderForEffect(run)
					return nil
				}),
		},
	})
}
