package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b9_helpers.go — shared bodies for the Commander deck cards of
// batches 9 to 12 (#5368).

// cmdrB9YouSacrificedLand is "whenever you sacrifice a land": the
// sacrifice event (CR 701.21) of a land its controller sacrificed. The
// event fires while the land is still on the battlefield, so it reads
// as a land (and a land token counts, as the printed text says "a land").
func cmdrB9YouSacrificedLand(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand()
}

// cmdrB9YouSacrificedAnotherNontoken is "whenever you sacrifice another
// nontoken permanent".
func cmdrB9YouSacrificedAnotherNontoken(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// cmdrB9YouSacrificedNontoken is "whenever you sacrifice a nontoken
// permanent" (the source itself included).
func cmdrB9YouSacrificedNontoken(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// cmdrB9SacrificedNontokenThisTurn is "as long as you've sacrificed a
// nontoken permanent this turn" (TurnTally.SacrificedCardTypes, kept at
// the sacrifice itself).
func cmdrB9SacrificedNontokenThisTurn(g *game.Game, you uuid.UUID) bool {
	return g.SacrificedWithCardTypeThisTurn(you, game.SacrificedNontokenKey) > 0
}

// cmdrB9ThenProliferate is "Proliferate." as the tail of a sentence that
// pauses on a choice: the proliferate waits for the answer. An on-disk key.
var cmdrB9ThenProliferate = CardTail("cmdr-b9/then-proliferate", func(ctx *Context, _ game.TailOutcome) error {
	return Proliferate{}.Apply(ctx)
})

// cmdrB9BasicLandCard is "basic land card": a land with the Basic supertype.
func cmdrB9BasicLandCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return IsBasicLand(c) }
}

// cmdrB9AuguryTake is Experimental Augury's look: one into your hand, the
// rest on the bottom in any order, then proliferate.
var cmdrB9AuguryTake = RegisterTake("choose-cards/experimental-augury-take",
	TakeDef{Then: cmdrB9AuguryBottomThenProliferate})

// cmdrB9AuguryBottomThenProliferate puts the rest on the bottom (the
// player orders them) and proliferates once that is settled.
func cmdrB9AuguryBottomThenProliferate(g *game.Game, res TakeFromLibraryResult) error {
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   res.Player,
		Source:    res.Source,
		Cards:     res.Rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    libraryOrderLabel(g, res.Source, "put the rest on the bottom of your library in any order"),
		ThenTail:  game.KeyedTail{Ref: cmdrB9ThenProliferate, Player: res.Player, Source: res.Source},
	})
}

// cmdrB9LandsOnly is a discard's set rule: every card chosen is a land card.
func cmdrB9LandsOnly(picked []game.Card) bool {
	for _, c := range picked {
		if !c.IsLand() {
			return false
		}
	}
	return true
}

// cmdrB9PermanentCard is "permanent card": a card with a permanent type.
func cmdrB9PermanentCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsPermanent()
	}
}
