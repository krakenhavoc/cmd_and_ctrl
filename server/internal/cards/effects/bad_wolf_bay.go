package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bad Wolf Bay — Plane — Earth (Doctor Who, #4870):
//
//	"At the beginning of combat on your turn, exile up to one target
//	 creature. Return it to the battlefield under its owner's control at
//	 the beginning of the next end step.
//	 When chaos ensues, cards can't enter from exile this turn. Then
//	 planeswalk."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability exiles the creature; its return is a delayed trigger
// (CR 603.7) that no longer depends on the plane, so it still happens
// after a planeswalk away. The chaos ability (CR 311.7) leaves a stored
// "this turn" rule (game.ModCardsCantEnterFromExile) that the one entry
// gate reads, so a card exiled by the first ability can't return this
// turn (CR 101.2: it stays in exile), and then the planar controller
// planeswalks at once as it resolves (CR 701.31, the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "706720f6-bb04-4a7b-8471-0f3d662d6eeb",
		Name:         "Bad Wolf Bay",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Bad Wolf Bay — exile up to one target creature; return it at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ExileTarget{Target: id, ThenTail: withCarry(ctx.ItemTail(badWolfBayExiled), id)}.Apply(ctx)
				}),
				TargetCreature("up to one target creature").WithCount(0, 1)),
			WheneverChaosEnsues("Bad Wolf Bay — cards can't enter from exile this turn; then planeswalk",
				func(g *game.Game, item *game.StackItem) error {
					g.CardsCantEnterFromExileThisTurnForEffect(item.SourceCardID, "Bad Wolf Bay — cards can't enter from exile this turn")
					if g.Planechase == nil {
						return nil
					}
					return g.PlaneswalkForEffect(item.Controller)
				}),
		},
	})
}

// badWolfBayExiled schedules the return once the creature reached exile.
// An on-disk key: never renamed, never reused.
var badWolfBayExiled = CardTail("bad-wolf-bay/return-at-next-end-step", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Bad Wolf Bay — return the exiled creature",
		Cards: []uuid.UUID{tailCarry0(o)},
		Body:  returnExiledToOwnersBody,
	}.Apply(ctx)
})
