package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Norn's Seedcore — Plane — New Phyrexia (Commander Masters, #4870):
//
//	"When you planeswalk to Norn's Seedcore, chaos ensues.
//	 Whenever chaos ensues, reveal cards from the top of your planar deck
//	 until you reveal a plane card. Planeswalk to it, except don't
//	 planeswalk away from any plane. Put the rest of the revealed cards on
//	 the bottom of your planar deck in any order."
//
// A plane (ADR 0216 §6). The planeswalk-to ability makes chaos ensue (CR
// 311.7), so every face-up plane's chaos ability triggers, including this
// one's (the ruling), and the planar controller orders them. The chaos
// ability reveals from the planar deck until a plane card (phenomena are
// revealed past), planeswalks to it without leaving any plane, so the game
// goes on across several planes at once until the next planeswalk, which
// leaves them all (CR 901.11c), and puts the other revealed cards on the
// bottom of the planar deck in an order the planar controller chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab72ff80-738a-4468-aecf-5d806143f791",
		Name:         "Norn's Seedcore",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("Norn's Seedcore — chaos ensues", func(g *game.Game, item *game.StackItem) error {
				g.ChaosEnsuesForEffect(item.Controller, uuid.Nil)
				return nil
			}),
			WheneverChaosEnsues("Norn's Seedcore — reveal until a plane card; planeswalk to it without planeswalking away; put the rest on the bottom of your planar deck in any order",
				func(g *game.Game, item *game.StackItem) error {
					run, plane := g.RevealPlanarUntilPlaneForEffect(item.Controller, item.SourceCardID, "Norn's Seedcore — revealed until a plane card")
					if plane != uuid.Nil {
						if err := g.PlaneswalkToCardForEffect(item.Controller, plane); err != nil {
							return err
						}
					}
					var rest []uuid.UUID
					for _, id := range run {
						if id != plane {
							rest = append(rest, id)
						}
					}
					g.PutPlanarCardsOnBottomInAnyOrderForEffect(rest)
					return nil
				}),
		},
	})
}
