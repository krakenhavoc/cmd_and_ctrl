package effects

import (
	"errors"
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// G'raha Tia, Scion Reborn — Legendary Creature — Cat Wizard {W}{U}{B},
// 2/3:
//
//	"Lifelink
//	 Throw Wide the Gates — Whenever you cast a noncreature spell, you
//	 may pay X life, where X is that spell's mana value. If you do,
//	 create a 1/1 colorless Hero creature token and put X +1/+1
//	 counters on it. Do this only once each turn."
//
// Lifelink is printed data. The ability triggers on every noncreature
// spell you cast (CR 603.2); the "may" and the payment are made as it
// resolves (CR 603.5, 608.2d, MayChoice with a LifeCost), not when it
// goes on the stack. "Do this only once each turn" limits the DOING,
// as Pantlaza, Sun-Favored's does its discover: a trigger whose "may"
// is declined, or whose life cannot be paid, does nothing and leaves
// the turn's one use for a later spell. The use is therefore recorded
// (Game.NoteDoneThisTurn) only when the life is paid and the Hero made,
// and it is read when the ability triggers (so no trigger is put on the
// stack once it is spent) and again as it resolves (so the second of
// two triggers stacked before either resolved does nothing).
//
// X is Monstrous Vortex's "that spell's mana value", read as the
// ability resolves, so an X the spell was cast for counts (CR 202.3e)
// and a spell countered in response is read where it went. Paying the
// life is a payment (CR 119.4): a player with less than X life is not
// asked, and the payment is re-checked after a "Yes". X of 0 is a
// payment of nothing and still makes the 1/1. The counters arrive as
// the token enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f558c758-8f12-45f9-8683-eee33ba1fb0d",
		Name:            "G'raha Tia, Scion Reborn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AllOf(YouCast(Noncreature()), grahaTiaNotYetDone), grahaTiaLabel, grahaTiaThrowWideTheGates),
		},
	})
}

const (
	grahaTiaLabel = "G'raha Tia — Throw Wide the Gates: you may pay X life to create a Hero with X +1/+1 counters"
	// grahaTiaDoneLabel is the tally the once-each-turn limit reads.
	grahaTiaDoneLabel = "G'raha Tia — Throw Wide the Gates (done)"
)

// grahaTiaNotYetDone keeps the trigger off the stack once the turn's
// one use has been spent.
func grahaTiaNotYetDone(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && g.ResolvedThisTurn(source.InstanceID, grahaTiaDoneLabel) == 0
}

// grahaTiaThrowWideTheGates asks, as the ability resolves, whether to
// pay X life, and makes the Hero only if that was paid.
func grahaTiaThrowWideTheGates(g *game.Game, item *game.StackItem) error {
	if g.ResolvedThisTurn(item.SourceCardID, grahaTiaDoneLabel) > 0 {
		return nil
	}
	x := triggeringSpellManaValue(g, item)
	if !grahaTiaCanPay(g, item, x) {
		return nil
	}
	return MayChoice{
		Question: "G'raha Tia — pay X life, where X is that spell's mana value, to create a 1/1 Hero with X +1/+1 counters?",
		LifeCost: x,
		Then:     grahaTiaAnswer,
		Word:     strconv.Itoa(x),
	}.Apply(NewContext(g, item))
}

func grahaTiaCanPay(g *game.Game, item *game.StackItem, x int) bool {
	p := g.PlayerByIDForEffect(item.Controller)
	return p != nil && !p.Eliminated && p.Life >= x
}

// grahaTiaPayAndCreate is the "Yes" branch: the payment is re-checked
// (CR 119.4), and only a payment that is made makes the Hero and uses
// up the turn.
func grahaTiaPayAndCreate(ctx *Context, x int) error {
	g, item := ctx.Game, ctx.Item
	if g.ResolvedThisTurn(item.SourceCardID, grahaTiaDoneLabel) > 0 || !grahaTiaCanPay(g, item, x) {
		return nil
	}
	if err := g.PayLifeForEffect(item.SourceCardID, item.Controller, x); err != nil {
		if errors.Is(err, game.ErrInvalidParam) {
			return nil // a payment that can't be made is not made (CR 119.8)
		}
		return err
	}
	g.NoteDoneThisTurn(item.SourceCardID, grahaTiaDoneLabel)
	opts := game.TokenEntryOptions{}
	if x > 0 {
		opts.Counters = map[string]int{game.CounterPlusOne: x}
	}
	_, err := g.CreateTokensForEffect(item.Controller, NewContext(g, item).TextToken(TokenCard("1/1 colorless Hero")), 1, opts)
	return err
}

// grahaTiaAnswer is the "yes": pay X life (Word) and create the Hero. Keyed
// (#4291).
var grahaTiaAnswer = MayYes("confirm/graha-tia-pay-and-create", func(ctx *Context, r game.ConfirmAnswer) error {
	return grahaTiaPayAndCreate(ctx, wordInt(r))
})
