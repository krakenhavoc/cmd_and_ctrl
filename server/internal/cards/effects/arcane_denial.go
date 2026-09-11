package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcane Denial — Instant {1}{U} (EDHREC rank 53):
//
//	"Counter target spell. Its controller may draw up to two cards
//	 at the beginning of the next turn's upkeep.
//	 You draw a card at the beginning of the next turn's upkeep."
//
// The politest counterspell in the format — the victim gets two
// cards for their trouble, which is why a table forgives it — and
// the highest-ranked card in the roadmap's batch 01.
//
// Both draws ride ONE CR 603.7 delayed trigger, scheduled for
// StepUpkeep with no owner gate: "the next turn's upkeep" is the
// very next upkeep at the table, whoever's it is. The card prints
// two abilities, but two simultaneous triggers under one controller
// queue a CR 603.3b ordering prompt (S19 sub-PR 8) and the order of
// two draws is unobservable — so one item, victim's draw first, is
// the same game with one fewer click every upkeep. It is controlled
// by Arcane Denial's controller (CR 603.7d) even though it draws for
// someone else first. A countered ability (no card on the stack)
// still schedules the caster's draw; a player who has left the
// table by then simply gets nothing.
//
// The victim's ID rides in the Effect closure BY VALUE. The
// no-capture rule on delayed triggers is about pointers into game
// state (a *Card, the *Game), which undo would leave dangling; a
// copied uuid is as clone-safe as the item's own Controller field,
// and DelayedTrigger.Cards cannot carry it — a player ID stamped as
// a TargetCard ref would fail the CR 608.2b existence re-check and
// fizzle the trigger.
//
// Sandbox simplification: "may draw UP TO two" is resolved as "draws
// two". There is no resolution-time number prompt for a trigger, and
// the only game in which a player declines a free draw is the one
// where their library is nearly empty — so this is weaker than
// printed for the caster in every ordinary game (the victim is never
// short-changed) and stronger only in the decking corner, which is
// declared here rather than hidden.
func init() {
	Register(Spec{
		OracleID: "ab1cc360-b9de-48d9-9983-4dfe4a7d2a37",
		Name:     "Arcane Denial",
		Targets:  TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			// Read the victim BEFORE countering: CounterTarget deletes
			// the StackMeta entry the controller lives on.
			victim := uuid.Nil
			if v := ctx.Game.StackItemForEffect(stackID); v != nil {
				victim = v.Controller
			}
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			return ScheduleDelayedTrigger{
				At:     game.StepUpkeep,
				Label:  "Arcane Denial — its controller draws two cards, you draw a card",
				Effect: arcaneDenialDraws(victim),
			}.Apply(ctx)
		},
	})
}

// arcaneDenialDraws builds the delayed trigger's effect: the
// countered spell's controller draws two, then the item's controller
// draws one. The victim is a copied ID, not a pointer, so the
// closure survives Clone / undo — see the card comment. A victim who
// has left the table, or who was never a player (a countered
// ability), is skipped.
func arcaneDenialDraws(victim uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if victim != uuid.Nil && g.PlayerByIDForEffect(victim) != nil {
			if err := (DrawCards{Player: victim, N: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
}
