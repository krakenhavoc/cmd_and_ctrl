package effects

import (
	"errors"
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_copy_life_paid.go — "whenever you activate an ability that
// isn't a mana ability, if life was paid to activate it, you may pay
// that much life again. If you do, copy that ability. You may choose
// new targets for the copy" (Verrak, Warped Sengir, #5662).
//
// The activation event says how much life the announcement paid
// (game.Event.LifePaid): a printed "Pay N life", what an X or counted
// life cost came to, and 2 for each Phyrexian mana symbol paid with
// life (CR 107.4f), in one total. A mana ability announces a different
// event kind and never reaches these (CR 605.3b).

// YouActivatedAnAbilityPayingLife is the When, on an
// EventActivateAbility: the source's controller activated the ability
// and paid life to do it. Paying 0 life is not paying life.
func YouActivatedAnAbilityPayingLife(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller && ev.LifePaid > 0
}

// LifeWasPaidToActivateIt is the CR 603.4 recheck of "if life was paid
// to activate it" as the trigger resolves. What was paid is fixed when
// the ability is activated, so it reads the triggering event, not the
// stack: the ability may have left it.
var LifeWasPaidToActivateIt = CondBoard{
	Name: "life was paid to activate that ability",
	Test: func(ctx *Context) bool { return ctx.Trigger().Event.LifePaid > 0 },
}

// MayPayThatLifeAgainToCopyIt is "you may pay that much life again. If
// you do, copy that ability. You may choose new targets for the copy."
// A player who can't pay that much life is not asked (CR 119.4,
// CR 119.8). The ability is copied only if the life is paid, and only
// while it is still on the stack, as every ability copy is
// (CopyAbility).
var MayPayThatLifeAgainToCopyIt Instruction = func(ctx *Context) error {
	life := ctx.Trigger().Event.LifePaid
	if life <= 0 || !canPayLifeNow(ctx, life) {
		return nil
	}
	question := "Pay " + strconv.Itoa(life) + " life again to copy that ability?"
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		question = c.Name + " — " + question
	}
	return MayChoice{
		Question: question,
		YesLabel: "Pay " + strconv.Itoa(life) + " life",
		NoLabel:  "Don't pay",
		LifeCost: life,
		Then:     mayPayLifeCopyThatAbility,
	}.Apply(ctx)
}

// mayPayLifeCopyThatAbility is the "yes": the payment is re-checked
// when the answer arrives (CR 119.4), and only a payment that is made
// copies the ability the triggering event named.
var mayPayLifeCopyThatAbility = MayYes("confirm/pay-life-copy-that-ability", func(ctx *Context, r game.ConfirmAnswer) error {
	if r.LifeCost <= 0 || !canPayLifeNow(ctx, r.LifeCost) {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), r.LifeCost); err != nil {
		if errors.Is(err, game.ErrInvalidParam) {
			return nil // a payment that can't be made is not made (CR 119.8)
		}
		return err
	}
	return CopyAbility{
		ItemID:           ctx.Trigger().Event.StackItemID,
		Controller:       ctx.Controller(),
		ChooseNewTargets: true,
	}.Apply(ctx)
})

// canPayLifeNow reports whether the resolving effect's controller can
// pay `life` right now: enough life (CR 119.4) and no "can't pay life"
// lock (CR 119.8).
func canPayLifeNow(ctx *Context, life int) bool {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	return p != nil && !p.Eliminated && ctx.Game.CanPayLifeLocked(p, life)
}
