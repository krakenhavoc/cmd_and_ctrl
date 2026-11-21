package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splatter Goblin — Creature — Phyrexian Goblin {1}{B}, 2/1 (Dominaria
// United):
//
//	"When this creature dies, target creature an opponent controls gets
//	 -1/-1 until end of turn."
//
// The target is chosen when the trigger goes on the stack; a target that is
// gone or no longer an opponent's creature by resolution is illegal and the
// shrink does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7cf5b74-123c-4def-a4aa-771ecd3be260",
		Name:         "Splatter Goblin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisDies("Splatter Goblin — target creature an opponent controls gets -1/-1 until end of turn",
					splatterGoblinShrink),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -1, -1))),
		},
	})
}

func splatterGoblinShrink(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range effects.LegalTargetCardIDs(ctx) {
		return effects.BoostUntilEOT{Target: id, Power: -1, Toughness: -1,
			Label: "Splatter Goblin — -1/-1 until end of turn"}.Apply(ctx)
	}
	return nil
}
