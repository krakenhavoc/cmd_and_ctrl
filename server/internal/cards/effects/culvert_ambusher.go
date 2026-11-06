package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Culvert Ambusher — Creature — Wurm Horror {3}{G}{G}, 4/5 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature enters or is turned face up, target creature
//	 blocks this turn if able.
//	 Disguise {4}{G} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// One ability with two conditions (OnAny, Sun Titan's shape). Entering
// face down does not trigger it, because a face-down permanent has no
// abilities (CR 708.2a); turning it face up does (CR 708.8). The block
// requirement is Peema Aether-Seer's CR 509.1c requirement until end
// of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c11fef0f-972d-4756-9902-fd3e3af00707",
		Name:             "Culvert Ambusher",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{4}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, Self,
					"Culvert Ambusher — target creature blocks this turn if able",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return BlockRequirementUntilEOT{
							Target: FirstLegalBattlefieldTarget(ctx),
							Kind:   game.BlockRequirementBlocks,
							Label:  "Culvert Ambusher — blocks this turn if able",
						}.Apply(ctx)
					}),
				TargetCreature("target creature")),
		},
	})
}
