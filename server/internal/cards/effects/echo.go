package effects

import (
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// echo.go — CR 702.30, ADR 0108 §5 (#1888).
//
//	702.30a Echo is a triggered ability. "Echo [cost]" means "At the
//	        beginning of your upkeep, if this permanent came under your
//	        control since the beginning of your last upkeep, sacrifice it
//	        unless you pay [cost]."
//
// One trigger built out of what exists: an AtYourUpkeep-shaped row, the
// engine's one fact for "came under your control since the beginning of
// your last upkeep" (game.CameUnderControlSinceLastUpkeepForEffect,
// game/echo.go), and the upkeep pay-unless prompt, which holds the
// table in the step until it is answered (#997).
//
//   - The "if" is an intervening if (CR 603.4): AppliesTo checks it as
//     the upkeep begins, and the effect checks it again as the trigger
//     resolves. A permanent that left (or left and came back, CR 400.7)
//     or changed controller in response is not charged.
//   - "Sacrifice it unless you pay" is CR 118.12a: the controller
//     chooses on resolution, and declining (or a "yes" they cannot
//     fund) sacrifices it.
//   - Paying emits EventEchoPaid, which Shah of Naar Isle's "When this
//     creature's echo cost is paid" watches.
//   - The cost is read on resolution, so Volcano Hellion's "echo {X},
//     where X is your life total" charges the life total the controller
//     has then (its ruling).
//
// Not a canonical keyword token, for cumulative upkeep's reason: echo
// carries a cost, and a bare token has nowhere to put it. The trigger
// carries game.KeywordEcho in TriggeredAbility.Keyword, which is how the
// table's "Echo due" marker and the coverage probe find it.

// UpkeepPayment is what an upkeep pay-unless charges: mana, or a payment
// that is not mana (ADR 0108 §5). Exactly one of Mana and Action is set.
type UpkeepPayment struct {
	// Mana is a mana cost ("{1}{G}").
	Mana string
	// Action is "discard N cards" or "sacrifice N permanents of a type".
	Action *game.PayAction
	// Words is how the prompt names an Action ("Discard a card",
	// "Sacrifice two lands"). Unused for Mana.
	Words string
}

// prompt is the payment's Cost string and Action for UpkeepPayUnless.
func (p UpkeepPayment) prompt() (string, *game.PayAction) {
	if p.Action != nil {
		return p.Words, p.Action
	}
	return p.Mana, nil
}

// DiscardPayment is "Discard N cards" as a payment.
func DiscardPayment(n int) UpkeepPayment {
	words := "Discard a card"
	if n != 1 {
		words = "Discard " + numberWord(n) + " cards"
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionDiscard, Count: n}, Words: words}
}

// SacrificePayment is "Sacrifice N <noun>" as a payment: `noun` and
// `plural` are the printed words ("land", "lands"), `of` the permanents
// that may pay.
func SacrificePayment(n int, noun, plural string, of game.PermanentQuery) UpkeepPayment {
	words := "Sacrifice " + articleFor(noun) + " " + noun
	if n != 1 {
		words = "Sacrifice " + numberWord(n) + " " + plural
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionSacrifice, Count: n, Of: of}, Words: words}
}

// times is the payment made `n` times over, all at once: a cumulative
// upkeep's cost for each age counter (CR 702.24a). A mana cost repeats
// its symbols; a discard or sacrifice multiplies its count.
func (p UpkeepPayment) times(n int, noun, plural string) UpkeepPayment {
	if p.Action == nil {
		return UpkeepPayment{Mana: strings.Repeat(p.Mana, n)}
	}
	count := p.Action.Count * n
	if p.Action.Kind == game.PayActionDiscard {
		return DiscardPayment(count)
	}
	return SacrificePayment(count, noun, plural, p.Action.Of)
}

// articleFor is "a" or "an" before a noun.
func articleFor(noun string) string {
	if noun != "" {
		switch noun[0] {
		case 'a', 'e', 'i', 'o', 'u':
			return "an"
		}
	}
	return "a"
}

// Echo is "Echo [cost]" for a mana cost on the card `name`:
// Echo("Acridian", "{1}{G}").
func Echo(name, cost string) game.TriggeredAbility {
	return EchoPaying(name+" — echo "+cost, func(*Context) UpkeepPayment { return UpkeepPayment{Mana: cost} })
}

// EchoPayment is "Echo—<payment>" for a fixed payment that is not mana:
// EchoPayment("Deepcavern Imp", DiscardPayment(1)) is "Echo—Discard a
// card".
func EchoPayment(name string, p UpkeepPayment) game.TriggeredAbility {
	return EchoPaying(name+" — echo: "+strings.ToLower(p.Words[:1])+p.Words[1:],
		func(*Context) UpkeepPayment { return p })
}

// EchoX is "echo {X}, where X is <where>", the amount read as the
// trigger resolves: Volcano Hellion's "where X is your life total". A
// negative amount is zero (CR 107.1b).
func EchoX(name, where string, amount func(ctx *Context) int) game.TriggeredAbility {
	return EchoPaying(name+" — echo {X}, where X is "+where, func(ctx *Context) UpkeepPayment {
		x := amount(ctx)
		if x < 0 {
			x = 0
		}
		return UpkeepPayment{Mana: "{" + strconv.Itoa(x) + "}"}
	})
}

// EchoPaying is the one echo trigger, with its payment computed as it
// resolves. `label` is the stack label.
func EchoPaying(label string, price func(ctx *Context) UpkeepPayment) game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword: game.KeywordEcho,
		Watches: []game.EventKind{game.EventBeginUpkeep},
		Key:     label,
		// CR 603.4, as the upkeep begins: your upkeep, and this
		// permanent came under your control since the beginning of
		// your last one.
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.Actor == source.Controller &&
				g.CameUnderControlSinceLastUpkeepForEffect(source.InstanceID, source.Controller)
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			source := item.SourceCardID
			// CR 603.4, again on resolution: the same object, still
			// yours, still charged. One that left, came back as a new
			// object (CR 400.7) or changed hands is not.
			if sourceIsNewObject(g, item) || !g.CameUnderControlSinceLastUpkeepForEffect(source, item.Controller) {
				return nil
			}
			ctx := NewContext(g, item)
			cost, action := price(ctx).prompt()
			return UpkeepPayUnless{
				Chooser:  item.Controller,
				Cost:     cost,
				Action:   action,
				Question: label + ": pay it, or sacrifice it?",
				OnDecline: func(ctx *Context) error {
					if sourceIsNewObject(ctx.Game, ctx.Item) {
						return nil
					}
					return SacrificeThisIfStillOnBattlefield(ctx.Game, ctx.Item)
				},
				OnPay: func(ctx *Context) error {
					ctx.Game.AnnounceEchoPaidForEffect(source, ctx.Item.Controller)
					return nil
				},
			}.Apply(ctx)
		},
	}
}

// WhenEchoIsPaid is "When this creature's echo cost is paid" (Shah of
// Naar Isle).
func WhenEchoIsPaid(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventEchoPaid, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.CardID == source.InstanceID
	}, label, effect)
}
