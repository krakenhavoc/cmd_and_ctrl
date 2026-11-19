package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hamlet Captain — Creature — Human Warrior {1}{G}, 2/2 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Whenever this creature attacks or blocks, other Humans you control
//	 get +1/+1 until end of turn."
//
// One printed ability that watches two events (Atlas's shape). The
// Humans are the ones you control as it resolves (CR 611.2c), never the
// Captain itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a60dc0f-e387-4d28-8ff0-c3f1883d0c5d",
		Name:         "Hamlet Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
				},
				"Hamlet Captain — other Humans you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					self := ctx.Source()
					match := And(Creature(), YouControl(), HasSubtype("Human"),
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
					return BoostUntilEOT{Match: match, Power: 1, Toughness: 1, Label: "Hamlet Captain — +1/+1"}.Apply(ctx)
				}),
		},
	})
}
