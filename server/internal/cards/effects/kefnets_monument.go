package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kefnet's Monument —
//
// "Blue creature spells you cast cost {1} less to cast. Whenever you cast a
// creature spell, target creature an opponent controls doesn't untap during
// its controller's next untap step."
func init() {
	Register(Spec{
		OracleID:     "b6294891-79e6-4f2a-a82d-6cffce968356",
		Name:         "Kefnet's Monument",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Blue creature spells you cast cost {1} less to cast", YourSpell(), func(q game.CostQuery) bool {
				return q.Card.IsCreature() && q.Card.HasColor("U")
			}),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, YouCast(Creature()), "Kefnet's Monument — target creature doesn't untap",
				func(g *game.Game, item *game.StackItem) error {
					return b751FreezeTarget(NewContext(g, item), uuid.Nil)
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
