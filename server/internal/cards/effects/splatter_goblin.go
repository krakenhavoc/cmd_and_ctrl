package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a7cf5b74-123c-4def-a4aa-771ecd3be260",
		Name:         "Splatter Goblin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisDies("Splatter Goblin — target creature an opponent controls gets -1/-1 until end of turn",
					splatterGoblinShrink),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -1, -1))),
		},
	})
}

func splatterGoblinShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCardIDs(ctx) {
		return BoostUntilEOT{Target: id, Power: -1, Toughness: -1,
			Label: "Splatter Goblin — -1/-1 until end of turn"}.Apply(ctx)
	}
	return nil
}
