package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Timbermare — Creature — Elemental Horse, {3}{G}, 5/5:
//
//	"Haste
//	 Echo {5}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, tap all other creatures."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e25e4f5-5676-480f-a52b-43ebb6f31537",
		Name:            "Timbermare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Echo("Timbermare", "{5}{G}"),
			WhenThisEnters("Timbermare — tap all other creatures", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var ids []uuid.UUID
				for _, c := range MatchingBattlefield(ctx, And(Creature(), NotSelf(item.SourceCardID))) {
					ids = append(ids, c.InstanceID)
				}
				return tapEach(ctx, ids)
			}),
		},
	})
}
