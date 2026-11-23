package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace, Reality Sculptor — Legendary Planeswalker — Jace {3}{U}{U},
// loyalty 5 (Reality Fracture):
//
//	"+1: Empower Jace X, where X is the number of Islands you control.
//	 −3: Until your next turn, whenever a creature attacks you or a
//	     planeswalker you control, it gets -5/-0 until end of turn.
//	 0: Exile all but the bottom card of each opponent's library.
//	    Activate only if there are twenty-five or more loyalty
//	    counters among Jaces you control."
//
// ADR 0139 proof card for all three of the seam's numbers:
//
//   - The +1 is "Empower Jace X" with X counted as the ability resolves
//     (EmpowerJace.Count). Jace himself is a Jace but not a token, so
//     the counters go on a Jace TOKEN — created by the ability if there
//     is none — never on him.
//   - The −3 is a CR 603.7b delayed trigger with a stated duration
//     ("until your next turn"), watching attack declarations: every
//     creature declared as an attacker against you or a planeswalker
//     you control gets -5/-0 until end of turn, through the stack.
//   - The 0 is gated by JaceLoyaltyAmong, "the number of loyalty
//     counters among Jaces you control": every Jace permanent you
//     control, him and the tokens alike (CR 602.1b, an activation
//     restriction, so the button is greyed and the bots never offer it
//     until it holds).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf0e9e00-fba5-448a-9052-ae730aea4bf8",
		Name:         "Jace, Reality Sculptor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Empower Jace X, where X is the number of Islands you control.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.EmpowerJace{Count: func(ctx *effects.Context) int {
						return islandsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Island"))
					}}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:  "−3: Until your next turn, whenever a creature attacks you or a planeswalker you control, it gets -5/-0 until end of turn.",
				Cost:   effects.LoyaltyCost(-3),
				Effect: jaceRealitySculptorMinusThree,
			},
			{
				Label: "0: Exile all but the bottom card of each opponent's library. Activate only if there are twenty-five or more loyalty counters among Jaces you control.",
				Cost:  effects.LoyaltyCost(0),
				Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
					return effects.JaceLoyaltyAmong(g, controller) >= 25
				},
				Effect: jaceRealitySculptorZero,
			},
		},
	})
}

// islandsControlledBy is "the number of Islands you control": any
// permanent with the Island land type, basic or not. `island` is the
// word as Jace's text now reads it (CR 612.2). Caller holds g.mu.
func islandsControlledBy(g *game.Game, player uuid.UUID, island string) int {
	g.RecomputeLayersIfStaleLocked()
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.HasSubtype(island) {
			n++
		}
	}
	return n
}

// jaceRealitySculptorMinusThree schedules the "until your next turn,
// whenever" trigger. Its controller is the ability's controller
// (CR 603.7d), and it is data — a registered condition and body — so a
// table with it waiting is still a restore point.
func jaceRealitySculptorMinusThree(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	controller := ctx.Controller()
	until := effects.DurationUntilYourNextTurn(ctx, controller)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        "Jace, Reality Sculptor — the attacking creature gets -5/-0 until end of turn",
		On:           []game.EventKind{game.EventAttack},
		Condition:    effects.JaceRealitySculptorAttacksYouCondition,
		Body:         jaceRealitySculptorShrinkBody,
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

// jaceRealitySculptorShrinkBody gives the attacking creature (the
// trigger's payload) -5/-0 until end of turn. A creature that has left
// the battlefield by then gets nothing.
var jaceRealitySculptorShrinkBody = game.SimpleDelayedBody("jace-reality-sculptor/attacker-gets-minus-five", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if _, ok := g.PermanentRefForEffect(id); !ok {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: id, Power: -5, Label: "Jace, Reality Sculptor — -5/-0"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// jaceRealitySculptorZero exiles every card but the bottom one of each
// opponent's library, one simultaneous exile per library. Index 0 is
// the bottom (Zone.Bottom).
func jaceRealitySculptorZero(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller || len(p.Library.Cards) < 2 {
			continue
		}
		ids := make([]uuid.UUID, 0, len(p.Library.Cards)-1)
		for _, c := range p.Library.Cards[1:] {
			ids = append(ids, c.InstanceID)
		}
		g.ExileCardsForEffect(ids)
	}
	return nil
}
