package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Insufferable Balladeer — Creature — Dwarf Bard {1}{R}, 2/1 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Vicious Mockery — When this creature enters, target creature an opponent
//	 controls can't block this turn. Goad it. (Until your next turn, that creature
//	 attacks each combat if able and attacks a player other than you if able.)"
//
// Ahn-Crop Crasher's can't-block restriction until end of turn, then the shared goad
// (CR 701.15a), both on the one target re-checked as the trigger resolves (CR 608.2b).
// "Vicious Mockery" is an ability word with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	label := "Insufferable Balladeer — target creature an opponent controls can't block this turn; goad it"
	effects.Register(effects.Spec{
		OracleID:     "e6acafdf-e3ad-4ac2-bb26-cec37c7ca115",
		Name:         "Insufferable Balladeer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				if err := effects.TargetCreatureCantBlockThisTurn("Insufferable Balladeer — can't block")(g, item); err != nil {
					return err
				}
				return effects.GoadTarget(g, item)
			}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
