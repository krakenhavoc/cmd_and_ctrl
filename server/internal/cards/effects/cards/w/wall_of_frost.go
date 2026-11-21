package cardsw

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Frost —
//
// "Defender Whenever this creature blocks a creature, that creature doesn't
// untap during its controller's next untap step."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "741e4f32-0587-40fa-a73d-5bcf66b52348",
		Name:            "Wall of Frost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlock},
			Key:     "Wall of Frost — attacking creature doesn't untap",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{item.Trigger.Event.Target}, Label: "Wall of Frost"}).Apply(effects.NewContext(g, item))
			}}},
	})
}
