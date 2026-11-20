package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "c11fef0f-972d-4756-9902-fd3e3af00707",
		Name:             "Culvert Ambusher",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{4}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, effects.Self,
					"Culvert Ambusher — target creature blocks this turn if able",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.BlockRequirementUntilEOT{
							Target: effects.FirstLegalBattlefieldTarget(ctx),
							Kind:   game.BlockRequirementBlocks,
							Label:  "Culvert Ambusher — blocks this turn if able",
						}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
