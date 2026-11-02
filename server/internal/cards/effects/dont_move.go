package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Don't Move — Sorcery {3}{W}{W} (Marvel Universe, #4460):
//
//	"Destroy all tapped creatures. Until your next turn, whenever a
//	 creature becomes tapped, destroy it."
//
// The sweep is DestroyAllMatching over tapped creatures, then a repeating
// delayed trigger (CR 603.7) lasting until the spell's controller's next
// turn begins, as Garruk, Curse Breaker's -4 does. It watches both tap
// events: an effect, cost or mana ability taps a creature with
// EventTapCard, but declaring an attacker taps it with only EventAttack,
// so an attacker that is tapped after the declaration counts and a
// vigilance attacker (which never becomes tapped) does not. Any
// creature's tapping triggers it, whoever controls it, the caster's
// included. The body destroys the creature that became tapped, if it is
// still on the battlefield as a creature (CR 603.7c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23cf095c-d43e-4a80-9bc5-f0f46ff885c4",
		Name:         "Don't Move",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DestroyAllMatching{Match: dontMoveTapped}).Apply(ctx); err != nil {
				return err
			}
			until := DurationUntilYourNextTurn(ctx, ctx.Controller())
			ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
				Controller:   ctx.Controller(),
				SourceCardID: ctx.Source(),
				Label:        "Don't Move — destroy the creature that became tapped",
				On:           []game.EventKind{game.EventTapCard, game.EventAttack},
				Condition:    dontMoveCreatureBecameTapped,
				Body:         dontMoveDestroyBody,
				Repeats:      true,
				Duration:     &until,
			})
			return nil
		},
	})
}

// dontMoveTapped is "tapped creatures".
func dontMoveTapped(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && c.Tapped
}

// dontMoveCreatureBecameTapped is "a creature becomes tapped": an
// EventTapCard of a creature, or an EventAttack whose creature is tapped
// after the declaration (the attack path emits no EventTapCard).
var dontMoveCreatureBecameTapped = game.DelayedCondition("tap/a-creature-becomes-tapped",
	func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventTapCard && ev.Kind != game.EventAttack {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || !c.IsCreature() {
			return false
		}
		return ev.Kind != game.EventAttack || c.Tapped
	})

// dontMoveDestroyBody destroys the creature the trigger saw become tapped.
var dontMoveDestroyBody = game.SimpleDelayedBody("dont-move/destroy-the-tapped-creature", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if _, ok := g.PermanentRefForEffect(id); !ok {
			continue
		}
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
