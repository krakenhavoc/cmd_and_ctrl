package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Together Forever — Enchantment {W}{W}:
//
//	"When this enchantment enters, support 2. (Put a +1/+1 counter on
//	 each of up to two target creatures.)
//	 {1}: Choose target creature with a counter on it. When that
//	 creature dies this turn, return that card to its owner's hand."
//
// The support is a permanent's, but its reminder text says "target
// creatures", not "other": the enchantment is not a creature, so the two
// read the same (SupportTargets with other=false).
//
// The activated ability schedules Whippoorwill's event-delayed trigger
// with the hand-return body: pinned to that creature object until
// cleanup, so a flicker ends it (CR 400.7), and controlled by the
// activator. "With a counter on it" is any kind, judged at announce and
// at resolution. The card returns only if it is still in the graveyard
// when the delayed trigger resolves (CR 603.7c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36b19ec0-d581-4213-bfae-1d7808a2f60d",
		Name:         "Together Forever",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersSupport("Together Forever", 2, false),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}: Choose target creature with a counter on it. When that creature dies this turn, return that card to its owner's hand.",
			Cost:    ManaCost("{1}"),
			Targets: TargetCreature("target creature with a counter on it", b41HasACounter()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				ReturnToHandWhenItDiesThisTurn(ctx, ts[0].ID, "Together Forever — return that card to its owner's hand")
				return nil
			},
		}},
	})
}

// ReturnToHandWhenItDiesThisTurn schedules "When <id> dies this turn,
// return that card to its owner's hand": ExileWhenItDiesThisTurn with
// the hand-return body.
func ReturnToHandWhenItDiesThisTurn(ctx *Context, id uuid.UUID, label string) {
	d := ctx.Game.PinnedTo(ctx.Game.UntilEndOfTurnDuration(), id)
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        label,
		On:           []game.EventKind{game.EventLTB},
		Condition:    theListedObjectDiedCondition,
		Cards:        []uuid.UUID{id},
		Duration:     &d,
		Body:         returnListedGraveyardToHandBody,
	})
}
