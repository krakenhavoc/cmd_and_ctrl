package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b049_helpers.go — shared seams of Strixhaven: School of Mages,
// batch 49 (#5502).

// devouringTendrilsGainBody is the fired trigger's body, "you gain 2
// life". An on-disk key: never renamed, never reused.
var devouringTendrilsGainBody = game.SimpleDelayedBody("old-b049/devouring-tendrils-gain-2", func(g *game.Game, item *game.StackItem) error {
	return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
})

// GainTwoLifeWhenItDiesThisTurn schedules "When <id> dies this turn, you
// gain 2 life" (Devouring Tendrils): an event-delayed trigger on the
// object, controlled by the resolving spell's controller, until cleanup
// and pinned to the object so a flicker ends it (CR 400.7). The object
// may be any permanent, not only a creature.
func GainTwoLifeWhenItDiesThisTurn(ctx *Context, id uuid.UUID, label string) {
	d := ctx.Game.PinnedTo(ctx.Game.UntilEndOfTurnDuration(), id)
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        label,
		On:           []game.EventKind{game.EventLTB},
		Condition:    TheListedObjectDiedCondition,
		Cards:        []uuid.UUID{id},
		Duration:     &d,
		Body:         devouringTendrilsGainBody,
	})
}
