package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcbound Shikari — Artifact Creature — Cat Soldier {1}{R}{W}, 0/0
// (Modern Horizons 2, #5501):
//
//	"First strike
//	 When this creature enters, put a +1/+1 counter on each other artifact
//	 creature you control.
//	 Modular 2 (...)"
//
// The set is read once as the trigger resolves; the Shikari gets none
// ("other"). Modular is the engine's keyword (CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "acb8d6a0-4bbd-4a42-bcac-3656367cc051",
		Name:            "Arcbound Shikari",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "modular 2"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arcbound Shikari — put a +1/+1 counter on each other artifact creature you control",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.IsArtifact() && c.InstanceID != item.SourceCardID {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.B13PutCounterOnEach(effects.NewContext(g, item), ids)
				}),
		},
	})
}
