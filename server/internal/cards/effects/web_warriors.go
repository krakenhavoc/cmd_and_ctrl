package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Web-Warriors — {4}{G/W} Creature — Spider Hero, 4/3 (Marvel's Spider-Man,
// #4460):
//
//	"When this creature enters, put a +1/+1 counter on each other creature
//	 you control."
//
// The set of creatures is fixed as the ability resolves (CR 608.2c), the
// source excluded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b23e4f8-882d-4cce-9b61-171ae2cc3a18",
		Name:         "Web-Warriors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Web-Warriors — a +1/+1 counter on each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != item.SourceCardID {
							ids = append(ids, c.InstanceID)
						}
					}
					return b13PutCounterOnEach(NewContext(g, item), ids)
				}),
		},
	})
}
