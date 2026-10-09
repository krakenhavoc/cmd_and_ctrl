package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk, Curse Breaker — Legendary Planeswalker — Garruk {3}{G}{G},
// loyalty 5 (Reality Fracture):
//
//	"Whenever a creature you control with power 4 or greater enters,
//	 draw a card.
//	 +2: Untap up to two target lands.
//	 −3: Create a 4/4 green Beast creature token with trample.
//	 −4: Until your next turn, whenever one or more creatures attack one
//	     of your opponents, those creatures get +2/+2 and gain trample
//	     until end of turn."
//
// The −4 is a CR 603.7b delayed trigger with a stated duration (Jace,
// Reality Sculptor's −3 is the template). The engine emits one attack
// event per attacking creature, so the trigger goes on the stack once
// per attacker and each resolution pumps that attacker; the net result
// is the same as one batched trigger over "those creatures". Only an
// attack on a PLAYER who is an opponent of Garruk's controller counts —
// an attack on a planeswalker is not an attack on the opponent.
func init() {
	Register(Spec{
		OracleID:     "184d4672-05a9-4182-8b3f-3561fb33ed62",
		Name:         "Garruk, Curse Breaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, garrukCurseBreakerBigEntered,
				"Garruk, Curse Breaker — draw a card",
				Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Untap up to two target lands.",
				Cost:    LoyaltyCost(2),
				Targets: TargetPermanent("up to two target lands", Land()).WithCount(0, 2),
				Effect:  b22UntapEachLegalTarget,
			},
			{
				Label:  "−3: Create a 4/4 green Beast creature token with trample.",
				Cost:   LoyaltyCost(-3),
				Effect: Do(CreateToken{Template: TokenCard("4/4 green Beast with trample"), N: 1}),
			},
			{
				Label:  "−4: Until your next turn, whenever one or more creatures attack one of your opponents, those creatures get +2/+2 and gain trample until end of turn.",
				Cost:   LoyaltyCost(-4),
				Effect: garrukCurseBreakerMinusFour,
			},
		},
	})
}

// garrukCurseBreakerBigEntered is "a creature you control with power 4
// or greater enters": power is read as the creature stands on the
// battlefield, counters and layers included.
func garrukCurseBreakerBigEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && c.CurrentPower() >= 4
}

func garrukCurseBreakerMinusFour(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	until := DurationUntilYourNextTurn(ctx, controller)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        "Garruk, Curse Breaker — the attacking creature gets +2/+2 and trample",
		On:           []game.EventKind{game.EventAttack},
		Condition:    garrukCurseBreakerAttacksOpponentCondition,
		Body:         garrukCurseBreakerPumpBody,
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

// garrukCurseBreakerAttacksOpponentCondition is "creatures attack one
// of your opponents": the declared defender (Event.Target) is a player
// other than the trigger's controller.
var garrukCurseBreakerAttacksOpponentCondition = game.DelayedCondition("attack/attacks-an-opponent-of-the-controller",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventAttack || ev.Target == uuid.Nil || ev.Target == dt.Controller {
			return false
		}
		return g.PlayerByIDForEffect(ev.Target) != nil
	})

var garrukCurseBreakerPumpBody = game.SimpleDelayedBody("garruk-curse-breaker/attacker-gets-plus-two-and-trample", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if _, ok := g.PermanentRefForEffect(id); !ok {
			continue
		}
		if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Garruk, Curse Breaker — +2/+2"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
