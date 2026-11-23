package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b24_helpers.go — helpers for BNS batch 24 (Eager Construct,
// Treasure Keeper).

// eagerConstructAsk asks each player in turn "scry 1?" (Model of Unity's
// ask-in-turn): a yes scries before the next player is asked.
func EagerConstructAsk(ctx *Context, players []uuid.UUID) error {
	if len(players) == 0 {
		return nil
	}
	return MayChoice{
		Player:   players[0],
		Question: "Eager Construct — scry 1?",
		Then:     eagerConstructAnswer,
		Carry:    players[1:],
	}.Apply(ctx)
}

// eagerConstructAnswer is one player's answer; either way the next player
// (the carried IDs) is asked after a yes-answer's scry. Registered in init
// because it asks the next player through eagerConstructAsk, which names it.
var eagerConstructAnswer game.ConfirmThen

func init() {
	eagerConstructAnswer = MayAnswer("confirm/eager-construct-scry", func(ctx *Context, r game.ConfirmAnswer) error {
		if !r.Accepted {
			return EagerConstructAsk(ctx, r.Carry)
		}
		next := ctx.ItemTail(eagerConstructAskNext)
		next.Carry = r.Carry
		return Scry{Player: r.Chooser, N: 1, ThenTail: next}.Apply(ctx)
	})
}

// eagerConstructAskNext asks the next player once the scry is done. An
// on-disk key.
var eagerConstructAskNext = CardTail("eager-construct/ask-next", func(ctx *Context, o game.TailOutcome) error {
	return EagerConstructAsk(ctx, o.Tail.Carry)
})

// treasureKeeperHit is "a nonland card with mana value 3 or less". A card
// whose cost the engine cannot read is not a hit (cascadeHit's rule), which
// errs on the weaker side.
func treasureKeeperHit(c game.Card) bool {
	if c.IsLand() {
		return false
	}
	if _, err := game.ParseCost(c.ManaCost); err != nil {
		return false
	}
	return c.ManaValue() <= 3
}

// treasureKeeperReveal is the dies trigger: reveal until a hit, offer its
// free cast from the library during this resolution (CR 608.2g), then put
// every revealed card not cast on the bottom in a random order.
func TreasureKeeperReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	run, hit := RevealUntil(ctx, you, treasureKeeperHit, "Treasure Keeper — reveal until a nonland card with mana value 3 or less")
	if len(run) == 0 {
		return nil
	}
	if hit == uuid.Nil {
		return g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, run)
	}
	perm := game.CastPermission{
		Player:     you,
		Zone:       game.ZoneLibrary,
		Duration:   g.UntilEndOfTurnDuration(),
		Cost:       "{0}",
		Timing:     game.TimingFlash,
		CastOnly:   true,
		Source:     ctx.Source(),
		SourceName: "Treasure Keeper",
		Label:      "Treasure Keeper — cast it without paying its mana cost",
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ctx.Source(),
		Reason:       "Treasure Keeper",
		Cards:        []uuid.UUID{hit},
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put it on the bottom",
		Then:         treasureKeeperBottomRest,
		Carry:        run,
	}) {
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, run)
}

// treasureKeeperBottomRest is "put all revealed cards not cast this way on
// the bottom of your library in a random order", once the cast is answered:
// Carry is every revealed card, and a card that was cast has left the
// library and is skipped. An on-disk key.
var treasureKeeperBottomRest = game.RegisterChooseCardsThen("resolution-cast/treasure-keeper-bottom-rest",
	func(g *game.Game, r game.CardsChosen) error {
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, r.Carry)
	})
