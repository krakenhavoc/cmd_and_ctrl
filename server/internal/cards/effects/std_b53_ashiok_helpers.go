package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b53_ashiok_helpers.go — the seams Ashiok, Wicked Manipulator needs
// (Standard batch 53, #4452): a replacement on paying life, and its
// Nightmare token.

// condCardPutIntoExileThisTurn is "if a card was put into exile this turn"
// (the Nightmare token): any card, from any zone, that moved to exile since
// the turn began, read off the turn's event log as the trigger fires and
// again as it resolves (CR 603.4). It declares no footprint condition
// (nothing in the trigger-independence vocabulary reads exile moves), so
// the row declares none.
type condCardPutIntoExileThisTurn struct{}

func (condCardPutIntoExileThisTurn) Holds(ctx *Context) bool {
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneExile {
			return true
		}
	}
	return false
}

func (condCardPutIntoExileThisTurn) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// Std53NightmareToken is Ashiok, Wicked Manipulator's 1/1 black Nightmare
// creature token with "At the beginning of combat on your turn, if a card
// was put into exile this turn, put a +1/+1 counter on this creature."
func Std53NightmareToken() game.Card { return tokenFromCatalog(printedStd53NightmareToken) }

// printedStd53NightmareToken is that Nightmare as PRINTED. "You" is the
// token's controller (item.Controller), and "this creature" the token.
// The slug is an on-disk identity: never renamed.
func printedStd53NightmareToken() tokenTemplate {
	return tokenTemplate{
		Slug: "nightmare-exile-counter",
		Card: game.Card{
			Name:      "Nightmare",
			TypeLine:  "Token Creature — Nightmare",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B"},
		},
		Triggered: []game.TriggeredAbility{
			InterveningIf(condCardPutIntoExileThisTurn{},
				On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
					"Nightmare — put a +1/+1 counter on this creature",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
		Text: "At the beginning of combat on your turn, if a card was put into exile this turn, put a +1/+1 counter on this creature.",
	}
}

// youExileFromLibraryInsteadOfPayingLife is "If you would pay life while
// your library has at least that many cards in it, exile that many cards
// from the top of your library instead" (Ashiok, Wicked Manipulator): a
// CR 614 replacement on a life PAYMENT (ReplacementEvent.LifePayment) by the
// source's controller. It is not optional and not partial (the rulings): a
// library with fewer cards than the payment leaves the payment alone, and
// the life is never split with the cards. The payment is replaced down to
// no life change (still paid, CR 614.17b) and the cards are exiled as the
// cost is paid, from the top, face up (an exile, not a mill, CR 701.17a).
func youExileFromLibraryInsteadOfPayingLife(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventLife || !ev.LifePayment || ev.LifeDelta >= 0 ||
				src == nil || src.Controller == uuid.Nil || ev.LifePlayer != src.Controller {
				return false
			}
			p := g.PlayerByIDForEffect(ev.LifePlayer)
			return p != nil && len(p.Library.Cards) >= -ev.LifeDelta
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			n := -ev.LifeDelta
			ev.LifeDelta = 0
			_, err := g.MillToZoneForEffect(ev.LifePlayer, n, game.ZoneExile)
			return err
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// ashiokTakeRestIntoExile is the continuation of Ashiok's +1: of the two
// cards looked at, the one taken is in hand and the other is exiled.
var ashiokTakeRestIntoExile = RegisterTake("choose-cards/ashiok-wicked-manipulator-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		var firstErr error
		for _, id := range res.Rest {
			if err := g.ExileCardForEffect(id); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}})
