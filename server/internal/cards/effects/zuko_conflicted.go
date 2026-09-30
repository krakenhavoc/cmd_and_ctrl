package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// zukoConflictedLabel is the trigger's stack label, and with it the key
// its "hasn't been chosen" memory is kept under.
const zukoConflictedLabel = "Zuko, Conflicted — beginning of your first main phase"

// Zuko, Conflicted — Legendary Creature — Human Rogue {B}{R}, 2/3:
//
//	"At the beginning of your first main phase, choose one that hasn't
//	 been chosen and you lose 2 life —
//	 • Draw a card.
//	 • Put a +1/+1 counter on Zuko.
//	 • Add {R}.
//	 • Exile Zuko, then return him to the battlefield under an
//	   opponent's control."
//
// ChooseOneNotChosen (ADR 0097): each bullet once for this object, and
// the fourth hands Zuko to an opponent as a new object (CR 400.7) with
// no memory, so the cycle starts again on their side of the table.
//
// "And you lose 2 life" is part of every choice, so it is the
// trigger's own body, which resolves before the chosen bullet (CR
// 608.2c runs the bullets after it). Once all four bullets are used a
// later trigger is removed with no effect (CR 700.2b) — and loses no
// life, because the whole instance is gone.
//
// "Add {R}" is mana from a trigger that is not a mana ability, so it
// arrives as the trigger resolves in the main phase and can be spent
// there. The first main phase is the precombat main phase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5fd82b9-77de-4358-9ce7-915cc809a889",
		Name:         "Zuko, Conflicted",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			zukoConflictedTrigger(),
		},
	})
}

func zukoConflictedTrigger() game.TriggeredAbility {
	t := AtYourPrecombatMain(zukoConflictedLabel, func(g *game.Game, item *game.StackItem) error {
		return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2)
	})
	t.Modes = ChooseOneNotChosen(
		ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		ModeDoing("Put a +1/+1 counter on Zuko.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		ModeDoing("Add {R}.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return AddMana{Player: item.Controller, Produced: "{R}"}.Apply(ctx)
			}),
		ModeDoing("Exile Zuko, then return him to the battlefield under an opponent's control.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return exileThisThenReturnUnderAnOpponentsControl(ctx, item, "Zuko, Conflicted")
			}),
	)
	return t
}
