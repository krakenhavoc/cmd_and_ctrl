package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zoraline, Cosmos Caller — Legendary Creature — Bat Cleric {1}{W}{B},
// 3/3 (Bloomburrow, #4467):
//
//	"Flying, vigilance
//	 Whenever a Bat you control attacks, you gain 1 life.
//	 Whenever Zoraline enters or attacks, you may pay {W}{B} and 2 life.
//	 When you do, return target nonland permanent card with mana value 3
//	 or less from your graveyard to the battlefield with a finality
//	 counter on it."
//
// Flying and vigilance are printed keywords. The Bat trigger fires once
// per attacking creature you control that is a Bat as it attacks, the
// word read through Zoraline's text (CR 612.2); Zoraline is a Bat
// herself.
//
// The enters-or-attacks trigger is Ripples of Undeath's mana-and-life
// offer: the {W}{B} rides the may-pay prompt and the 2 life is paid as a
// cost once the mana is, before anything else. The offer is not made to
// a player with less than 2 life (CR 119.4). A player whose life total
// can't change can't pay the life (CR 119.8), so nothing returns. Paying creates a
// CR 603.12 reflexive trigger, Terra, Herald of Hope's shape, whose
// target is chosen as it goes on the stack (the ruling). Its body is
// Grim Reaper's Scythe's return with a finality counter, which the
// engine's CR 122.1h replacement then reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a775f50-20f8-4a8f-868c-8907305bfad6",
		Name:            "Zoraline, Cosmos Caller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, zoralineBatAttacked, "Zoraline, Cosmos Caller — a Bat attacked: you gain 1 life", Do(GainLife{Amount: 1})),
			WhenThisEntersOrAttacks("Zoraline, Cosmos Caller — you may pay {W}{B} and 2 life to return a nonland permanent card with mana value 3 or less", zoralineOffer),
		},
	})
}

// zoralineLife is the life half of the payment.
const zoralineLife = 2

// zoralineBatAttacked is "a Bat you control attacks".
func zoralineBatAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Bat"))
}

// zoralineOffer asks for {W}{B} and 2 life, to a player with 2 life to pay.
func zoralineOffer(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Life < zoralineLife {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{W}{B}",
		Question: "Zoraline, Cosmos Caller — pay {W}{B} and 2 life to return a nonland permanent card with mana value 3 or less from your graveyard with a finality counter?",
		Then:     zoralinePaid,
	}.Apply(NewContext(g, item))
}

// zoralinePaid is the paid half, keyed (#4264): the 2 life is paid, then
// the reflexive trigger is queued against the asking item.
var zoralinePaid = PayUnlessBranches("pay-unless/zoraline-cosmos-caller-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), zoralineLife); err != nil {
			return nil // the life could not be paid, so there is no "when you do"
		}
		return ReflexiveTrigger{
			Label: "Zoraline, Cosmos Caller — return target nonland permanent card with mana value 3 or less from your graveyard with a finality counter",
			Body:  zoralineReturnBody,
		}.Apply(ctx)
	}, nil)

// zoralineReturnBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var zoralineReturnBody = game.ReflexiveBody("zoraline-cosmos-caller/return", simpleBody(grimReapersScytheReturn),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target nonland permanent card with mana value 3 or less from your graveyard",
			YouOwn(), NonlandPermanentCard(), ManaValueLE(3))
	}))
