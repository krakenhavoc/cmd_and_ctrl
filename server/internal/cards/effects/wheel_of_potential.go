package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wheel of Potential — Sorcery {2}{R} (Modern Horizons 3):
//
//	"You get {E}{E}{E} (three energy counters), then you may pay any amount
//	 of {E}.
//	 Each player may exile their hand and draw a number of cards equal to the
//	 amount of {E} paid this way. If seven or more {E} was paid this way, you
//	 may play cards you own exiled this way until the end of your next turn."
//
// The energy is Aether Spike's: GetEnergy, then the pay_amount prompt, whose
// answer (the amount paid, 0 when nothing was) is handed to a registered
// branch. The branch asks each player in APNAP order (MayChoice) whether to
// exile their hand and draw that many; a yes exiles the hand as one move and,
// once it has landed, grants the "play until the end of your next turn" window
// over every exiled card the spell's controller OWNS (only when seven or more
// was paid), then draws, then asks the next player. The window is the
// controller's alone: another player's exiled hand is gone for good, and a
// card the controller owns that was exiled from an opponent's hand is still
// playable (CR 108.3, "cards you own").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95a7db5d-26f6-4192-9dc7-4342f7e0e9c5",
		Name:         "Wheel of Potential",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3, Draws: 7},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GetEnergy{N: 3}).Apply(ctx); err != nil {
				return err
			}
			return PayEnergyAmount{
				Question: "Wheel of Potential — pay any amount of {E}; each player may exile their hand and draw that many cards",
				Unit:     game.PayAmountCards,
				Goal:     AsMuchAsYouCan,
				Then:     wheelOfPotentialPaid,
			}.Apply(ctx)
		},
	})
}

// wheelOfPotentialPaid is the paid half: the amount rides to every player's
// question. An on-disk key: never renamed, never reused.
var wheelOfPotentialPaid = NumberThen("number/wheel-of-potential-paid",
	func(ctx *Context, r game.NumberAnswer) error {
		return wheelOfPotentialAsk(ctx, apnapPlayers(ctx.Game), r.Number)
	})

// wheelOfPotentialAsk asks the first of `order` whether to exile their hand
// and draw n cards, and the rest afterwards.
func wheelOfPotentialAsk(ctx *Context, order []uuid.UUID, n int) error {
	for len(order) > 0 {
		who, rest := order[0], order[1:]
		if p := ctx.Game.PlayerByIDForEffect(who); p == nil || p.Eliminated {
			order = rest
			continue
		}
		return MayChoice{
			Player:   who,
			Question: "Wheel of Potential — exile your hand and draw " + strconv.Itoa(n) + " cards?",
			YesLabel: "Exile my hand, draw " + strconv.Itoa(n),
			NoLabel:  "Keep my hand",
			Then:     wheelOfPotentialAnswer,
			Carry:    rest,
			Word:     strconv.Itoa(n),
		}.Apply(ctx)
	}
	return nil
}

// wheelOfPotentialAnswer is one player's answer. On a yes their hand is
// exiled and the rest of the card runs from the exile's continuation; on a
// no the next player is asked. Keyed.
var wheelOfPotentialAnswer game.ConfirmThen

var wheelOfPotentialExiled game.TailRef

func init() {
	wheelOfPotentialAnswer = MayAnswer("confirm/wheel-of-potential-exile-hand", func(ctx *Context, r game.ConfirmAnswer) error {
		n := wordInt(r)
		if !r.Accepted {
			return wheelOfPotentialAsk(ctx, r.Carry, n)
		}
		tail := ctx.TailN(wheelOfPotentialExiled, n)
		tail.Carry = append([]uuid.UUID{r.Chooser}, r.Carry...)
		tail.Asked = game.NewAskedItem(ctx.Item)
		return ctx.Game.ExileCardsKeyedForEffect(allHandCardIDs(ctx.Game, r.Chooser), tail)
	})

	// wheelOfPotentialExiled runs once a hand has landed in exile: the controller's
	// play window over the cards they own (seven or more paid), the draw, then the
	// next player's question. The tail's Carry is the player who exiled, then the
	// players still to ask; N is the amount paid.
	wheelOfPotentialExiled = CardTail("wheel-of-potential/exiled-hand", func(ctx *Context, o game.TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		who, rest, n := o.Tail.Carry[0], o.Tail.Carry[1:], o.Tail.N
		me := ctx.Controller()
		if n >= 7 {
			for _, id := range o.IDs {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok || c.Owner != me {
					continue
				}
				ctx.Game.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
					Player:   me,
					Zone:     game.ZoneExile,
					Duration: ctx.Game.UntilEndOfYourNextTurnDuration(me),
				})
			}
		}
		if err := (DrawCards{Player: who, N: n}).Apply(ctx); err != nil {
			return err
		}
		return wheelOfPotentialAsk(ctx, rest, n)
	})
}
