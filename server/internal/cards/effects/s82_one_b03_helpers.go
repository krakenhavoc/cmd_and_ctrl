package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tapEachLegalTargetBody is "tap target creature" as a trigger body: every
// target still legal as the ability resolves is tapped (CR 608.2b).
func tapEachLegalTargetBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Ambulatory Edifice: "you may pay 2 life. When you do, target creature gets
// -1/-1 until end of turn." The answer pays the life (re-checked, CR 119.4)
// and raises the reflexive trigger (CR 603.12). On-disk keys: never renamed.
var ambulatoryEdificePaid = MayYes("confirm/ambulatory-edifice-pay-life", func(ctx *Context, r game.ConfirmAnswer) error {
	if !canPayLifeNow(ctx, 2) {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), 2); err != nil {
		if errors.Is(err, game.ErrInvalidParam) {
			return nil // a payment that can't be made is not made (CR 119.8)
		}
		return err
	}
	return ReflexiveTrigger{
		Label: "Ambulatory Edifice — target creature gets -1/-1 until end of turn",
		Body:  ambulatoryEdificeShrinkBody,
	}.Apply(ctx)
})

var ambulatoryEdificeShrinkBody = game.ReflexiveBody("ambulatory-edifice/shrink",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return BoostUntilEOT{Target: id, Power: -1, Toughness: -1,
			Label: "Ambulatory Edifice — -1/-1 until end of turn"}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

// Shrapnel Slinger: the sacrifice's "when you do".
var shrapnelSlingerSacrificed = RegisterSacrificedOne("prompt-run/shrapnel-slinger-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label: "Shrapnel Slinger — destroy target artifact an opponent controls",
		Body:  shrapnelSlingerDestroyBody,
	}.Apply(ctx)
})

var shrapnelSlingerDestroyBody = game.ReflexiveBody("shrapnel-slinger/destroy",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		return destroyTheTargetPermanent(item, NewContext(g, item))
	}),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target artifact an opponent controls", Artifact(), OpponentControls())
	}))

// Myr Custodian: "Then each opponent may scry 1", asked in turn order. A yes
// scries before the next opponent is asked. On-disk keys: never renamed.
func myrCustodianAsk(ctx *Context, opponents []uuid.UUID) error {
	if len(opponents) == 0 {
		return nil
	}
	return MayChoice{
		Player:   opponents[0],
		Question: "Myr Custodian — scry 1?",
		Then:     myrCustodianAnswer,
		Carry:    opponents[1:],
	}.Apply(ctx)
}

var myrCustodianAnswer game.ConfirmThen

func init() {
	myrCustodianAnswer = MayAnswer("confirm/myr-custodian-scry", func(ctx *Context, r game.ConfirmAnswer) error {
		if !r.Accepted {
			return myrCustodianAsk(ctx, r.Carry)
		}
		next := ctx.ItemTail(myrCustodianAskOpponents)
		next.Carry = r.Carry
		return Scry{Player: r.Chooser, N: 1, ThenTail: next}.Apply(ctx)
	})
}

var myrCustodianAskOpponents = CardTail("myr-custodian/ask-opponents", func(ctx *Context, o game.TailOutcome) error {
	return myrCustodianAsk(ctx, o.Tail.Carry)
})

// cruelGrimnarchEnters: each opponent discards a card as one run, then the
// controller gains 4 life for each opponent who discarded nothing (cruelGrimnarchGain).
func cruelGrimnarchEnters(g *game.Game, item *game.StackItem) error {
	return g.EachPlayerDiscardsThenKeyedForEffect(item.Controller,
		game.DiscardPrompt{Source: item.SourceCardID, N: 1},
		cruelGrimnarchGain, []uuid.UUID{item.Controller})
}

// cruelGrimnarchGain is "for each opponent who can't, you gain 4 life", run
// once the last discard has landed: an opponent with no card in hand
// discarded none (CR 701.9a). The controller is the carry. An on-disk key:
// never renamed, never reused.
var cruelGrimnarchGain = game.RegisterPromptRunThen("prompt-run/cruel-grimnarch-gain", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	you := r.Carry[0]
	discards := r.Discards()
	cant := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if len(discards.By(p.ID)) == 0 {
			cant++
		}
	}
	if cant == 0 {
		return nil
	}
	return GainLife{Player: you, Amount: 4 * cant}.Apply(keyedContext(g, you, r.Source))
})
