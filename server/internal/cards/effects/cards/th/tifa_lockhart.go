package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tifa Lockhart — Legendary Creature — Human Monk {1}{G}, 1/2 (Final
// Fantasy, #4462):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, double Tifa
//	 Lockhart's power until end of turn."
//
// Trample rides PrintedKeywords. "Double its power" is CR 701.10b: Tifa
// gets +X/+0 until end of turn, where X is her power as the trigger
// resolves, so two lands in a turn make her four times as strong. A Tifa
// that has left the battlefield, or came back as a new object (CR 400.7),
// is not doubled.
//
// No simplification.
func init() {
	const label = "Tifa Lockhart — double Tifa Lockhart's power until end of turn (landfall)"
	effects.Register(effects.Spec{
		OracleID:        "878df304-9d6d-4ce9-b65b-1439c251f3e0",
		Name:            "Tifa Lockhart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall(label, func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.DoublePowerUntilEOT(effects.NewContext(g, item), item.SourceCardID, "Tifa Lockhart — double power")
			}),
		},
	})
}
