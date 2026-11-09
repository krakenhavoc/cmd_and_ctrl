package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tymna the Weaver — Legendary Creature — Human Cleric {1}{W}{B}, 2/2
// (Final Fantasy, #4462):
//
//	"Lifelink
//	 At the beginning of each of your postcombat main phases, you may pay X
//	 life, where X is the number of opponents that were dealt combat damage
//	 this turn. If you do, draw X cards.
//	 Partner (You can have two commanders if both have partner.)"
//
// The trigger fires at the start of every postcombat main phase of yours
// (an added combat's included, CR 505.1a). X is read as it resolves, from
// the turn's record of which creatures dealt combat damage to which player:
// an opponent dealt combat damage by any creature counts, one who was dealt
// it and then lost the game still does (the ruling), and each counts once
// however much was dealt. You pay all of X or none (the ruling), re-checked
// as you answer: a player with less life than X can't pay it (CR 119.4).
// The draw follows the payment. X of 0 does nothing.
//
// Partner is a deck-construction rule and needs nothing on the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d15642e4-e61c-4d29-af48-de837991245e",
		Name:            "Tymna the Weaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			AtYourPostcombatMain("Tymna the Weaver — you may pay X life, where X is the number of opponents dealt combat damage this turn; if you do, draw X cards", tymnaOffer),
		},
	})
}

// tymnaOpponentsDealtCombatDamage is X: the opponents of `you` (a player
// who has since lost still counts) any creature dealt combat damage to this
// turn.
func tymnaOpponentsDealtCombatDamage(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.ID == you {
			continue
		}
		if len(g.CreaturesThatDealtCombatDamageToThisTurn(p.ID)) > 0 {
			n++
		}
	}
	return n
}

// tymnaOffer asks for the payment.
func tymnaOffer(g *game.Game, item *game.StackItem) error {
	x := tymnaOpponentsDealtCombatDamage(g, item.Controller)
	if x <= 0 {
		return nil
	}
	return MayChoice{
		Question: "Tymna the Weaver — pay " + strconv.Itoa(x) + " life to draw " + strconv.Itoa(x) + " cards?",
		YesLabel: "Pay " + strconv.Itoa(x) + " life",
		NoLabel:  "Don't pay",
		LifeCost: x,
		Draws:    x,
		Then:     tymnaAnswer,
		Word:     strconv.Itoa(x),
	}.Apply(NewContext(g, item))
}

// tymnaAnswer pays the life and draws that many cards. Keyed: X rides the
// prompt. An on-disk key: never renamed, never reused.
var tymnaAnswer = MayYes("confirm/tymna-the-weaver-pay-and-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	x := wordInt(r)
	me := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(me)
	if x <= 0 || p == nil || p.Eliminated || p.Life < x {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), me, x); err != nil {
		return err
	}
	return DrawCards{N: x}.Apply(ctx)
})
