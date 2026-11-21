package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aphelia, Viper Whisperer — Legendary Creature — Gorgon Assassin {1}{B},
// 1/3 (Foundations Jumpstart):
//
//		"Deathtouch
//		 Whenever Aphelia attacks, you may pay {1}{B/G}. If you do, create a
//		 1/1 black Snake creature token with deathtouch.
//		 {4}{B}: Until end of turn, whenever one or more Gorgons and/or Snakes
//		 you control deal combat damage to a player, that player loses half
//		 their life, rounded up."
//
//	  - The attack trigger is MayPay: the Snake is made only when the mana was
//	    really paid.
//	  - The activation starts a repeating delayed trigger for the turn (CR
//	    603.7b), an each-condition naming the damaged player, once per player
//	    per damage batch (CR 603.2c): "one or more" is one occurrence however
//	    many Gorgons and Snakes connect, and two players hit at once are two.
//	    Each activation is its own trigger, so activating twice halves twice.
//	    The creature's type is read as the damage is dealt, and "you control"
//	    is the activation's controller.
//	  - The loss is read as it resolves: half the player's life then, rounded
//	    up, and nothing at 0 or less.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "191e003f-1a78-4d85-9c70-2ae132a5f9cc",
		Name:            "Aphelia, Viper Whisperer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self,
				"Aphelia, Viper Whisperer — you may pay {1}{B/G}; if you do, create a 1/1 black Snake with deathtouch",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{B/G}",
						Question: "Aphelia, Viper Whisperer — pay {1}{B/G} to create a 1/1 black Snake creature token with deathtouch?",
						Then:     apheliaPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}: Until end of turn, whenever one or more Gorgons and/or Snakes you control deal combat damage to a player, that player loses half their life, rounded up.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				until := effects.DurationUntilEndOfTurn(ctx)
				g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
					Controller:   ctx.Controller(),
					SourceCardID: ctx.Source(),
					Label:        "Aphelia, Viper Whisperer — that player loses half their life, rounded up",
					On:           []game.EventKind{game.EventDealDamage},
					Condition:    apheliaGorgonOrSnakeHitsAPlayer,
					Body:         apheliaHalfLifeBody,
					Repeats:      true,
					OncePerBatch: true,
					Duration:     &until,
				})
				return nil
			},
		}},
	})
}

var apheliaPaid = effects.PayUnlessBranches("aphelia-viper-whisperer/paid", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	return effects.CreateToken{Template: effects.TokenCard("1/1 black Snake with deathtouch"), N: 1}.Apply(ctx)
}, nil)

// apheliaGorgonOrSnakeHitsAPlayer names the damaged player when a Gorgon or
// Snake the delayed trigger's controller controls dealt combat damage to them.
var apheliaGorgonOrSnakeHitsAPlayer = game.DelayedEachCondition("damage/gorgons-or-snakes-combat-to-player",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
			return nil
		}
		c, ok := g.LookupCardForEffect(ev.Source)
		if !ok || !c.IsCreature() || c.Controller != dt.Controller {
			return nil
		}
		// The words are Aphelia's text (CR 612.2); a source that has left
		// reads them as printed.
		reader, _ := g.LookupCardForEffect(dt.SourceCardID)
		if !c.HasSubtype(reader.TextSubtype("Gorgon")) && !c.HasSubtype(reader.TextSubtype("Snake")) {
			return nil
		}
		return []uuid.UUID{ev.Target}
	})

var apheliaHalfLifeBody = game.SimpleDelayedBody("life/aphelia-half-their-life", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Trigger().Event.Target
	p := ctx.PlayerByID(player)
	if p == nil || p.Life <= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(ctx.Source(), player, -((p.Life + 1) / 2))
})
