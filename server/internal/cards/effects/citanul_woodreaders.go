package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Citanul Woodreaders — Creature — Human Druid {2}{G}, 1/4 (Planar Chaos, #5595):
//
//	"Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)
//	 When this creature enters, if it was kicked, draw two cards."
//
// Battlewing Mystic's kicked enters trigger: ThisKickedAtLeast reads the
// kick off the spell that became this permanent, so an unkicked cast, or a
// Woodreaders put onto the battlefield, draws nothing (CR 702.33d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a4555bd2-be42-442a-8ead-3398318e91c5",
		Name:          "Citanul Woodreaders",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Citanul Woodreaders — kicked: draw two cards", Do(DrawCards{N: 2})),
		},
	})
}
