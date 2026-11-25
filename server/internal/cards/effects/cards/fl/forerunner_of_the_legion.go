package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forerunner of the Legion — Creature — Vampire Knight {2}{W}, 2/2:
//
//	"When this creature enters, you may search your library for a Vampire card,
//	 reveal it, then shuffle and put that card on top.
//	 Whenever another Vampire you control enters, target creature gets +1/+1
//	 until end of turn."
//
// The first ability is the Harbinger cycle's. The second triggers for each
// other Vampire entering under its controller's control and targets as it goes
// on the stack (CR 603.3d); a target that has left or stopped being a creature
// is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "439537ca-53c1-47cb-9174-5872c3c46288",
		Name:         "Forerunner of the Legion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Forerunner of the Legion", "a Vampire card", effects.LrwB1CardOfType("Vampire")),
			effects.Targeting(effects.On(game.EventETB, effects.S82OldB216TribeEnters("Vampire", true),
				"Forerunner of the Legion — target creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.LegalTargetCardIDs(ctx) {
						return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Forerunner of the Legion — +1/+1"}.Apply(ctx)
					}
					return nil
				}), effects.TargetCreature("target creature")),
		},
	})
}
