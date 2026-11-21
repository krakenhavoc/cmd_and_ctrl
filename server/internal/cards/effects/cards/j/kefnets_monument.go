package cardsj

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kefnet's Monument —
//
// "Blue creature spells you cast cost {1} less to cast. Whenever you cast a
// creature spell, target creature an opponent controls doesn't untap during
// its controller's next untap step."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6294891-79e6-4f2a-a82d-6cffce968356",
		Name:         "Kefnet's Monument",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Blue creature spells you cast cost {1} less to cast", effects.YourSpell(), func(q game.CostQuery) bool {
				return q.Card.IsCreature() && q.Card.HasColor(q.Source.TextColor("U"))
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, effects.YouCast(effects.Creature()), "Kefnet's Monument — target creature doesn't untap",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B751FreezeTarget(effects.NewContext(g, item), uuid.Nil)
				}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
