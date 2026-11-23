package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arni Metalbrow — Legendary Creature — Human Berserker {2}{R}, 3/3 (March
// of the Machine: The Aftermath):
//
//	"Whenever a creature you control attacks or enters attacking, you may
//	 pay {1}{R}. If you do, you may put a creature card with mana value
//	 less than that creature's mana value from your hand onto the
//	 battlefield tapped and attacking."
//
// "Attacks or enters attacking" is ONE printed trigger, so it is one
// registered trigger over two events: the attack declaration (one trigger
// per attacker, as Swordsman, Sharp Scoundrel's) and a creature entering
// the battlefield already attacking (Card.AttackingTarget set as it
// enters, CR 508.4: it was never declared, so it does not "attack" and
// only this half sees it). A creature that enters and is not attacking
// does nothing. {1}{R} is MayPay's optional payment; "if you do" is its
// paid branch, which reads the triggering creature's mana value (its last
// known value if it has since left, CR 608.2h) and offers Kinscaer
// Sentry's hand put (HandAttackerPutter) over the creature cards in hand
// with a strictly lower mana value. A card with {X} in its cost counts as
// X = 0. The put creature enters tapped and attacking, never declared, so
// it triggers nothing that watches attacks and cannot loop through this
// ability's attack half; it does trigger the enters-attacking half, as
// the card says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d799e58a-0032-4df8-a77f-0d824ec15214",
		Name:         "Arni Metalbrow",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventETB}, arniAttacksOrEntersAttacking,
				"Arni Metalbrow — you may pay {1}{R}; if you do, you may put a creature card with lesser mana value from your hand onto the battlefield tapped and attacking",
				arniMetalbrowTrigger),
		},
	})
}

func arniAttacksOrEntersAttacking(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == uuid.Nil || source == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || c.Controller != source.Controller {
		return false
	}
	switch ev.Kind {
	case game.EventAttack:
		return ev.Actor == source.Controller
	case game.EventETB:
		return c.AttackingTarget != uuid.Nil
	}
	return false
}

var arniPutter = effects.NewHandAttackerPutter("arni-metalbrow", "Arni Metalbrow", arniEntered)

// arniEntered follows the entry with nothing. An on-disk key.
var arniEntered = effects.EnteredTail("arni-metalbrow/entered", func(_ *effects.Context, _ uuid.UUID, _ game.TailOutcome) error {
	return nil
})

var arniPaid = effects.PayUnlessBranches("arni-metalbrow/paid", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	g := ctx.Game
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	mv := 0
	if c, ok := g.LookupCardForEffect(ctx.Item.Trigger.Event.CardID); ok {
		mv, _ = g.ManaValueForEffect(c)
	}
	arniPutter.Offer(g, ctx.Item, effects.HandCardsMatching(g, r.Chooser, effects.And(effects.Creature(), effects.ManaValueLE(mv-1))),
		"Arni Metalbrow — you may put a creature card with lesser mana value from your hand onto the battlefield tapped and attacking")
	return nil
}, nil)

func arniMetalbrowTrigger(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{R}",
		Question: "Arni Metalbrow — pay {1}{R} to put a creature card with lesser mana value from your hand onto the battlefield tapped and attacking?",
		Then:     arniPaid,
	}.Apply(effects.NewContext(g, item))
}
