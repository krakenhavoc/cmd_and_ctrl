package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB005CSilverCounter is the "silver" counter Karn, Scion of Urza puts
// on the card he exiles (a counter name is a plain string on a card).
const S82OldB005CSilverCounter = "silver"

// S82OldB005CKarnScionPlusOne is Karn, Scion of Urza's +1: "Reveal the top two
// cards of your library. An opponent chooses one of them. Put that card into
// your hand and exile the other with a silver counter on it." With several
// opponents the controller names the one who chooses, as every "an opponent
// chooses" clause of the catalog does (CR 101.4 does not apply: one chooser).
func S82OldB005CKarnScionPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{N: 2, Reason: "Karn, Scion of Urza — reveal the top two cards", Revealed: &revealed}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Karn, Scion of Urza — choose an opponent to choose one of the revealed cards",
		ThenKey:  s82OldB005CKarnChooser,
		Carry:    revealed,
	}.Apply(ctx)
}

var s82OldB005CKarnChooser = ChoosePlayerThen("option-pick/s82-old-b005-karn-scion-chooser",
	func(ctx *Context, chosen uuid.UUID, carry []uuid.UUID) error {
		if chosen == uuid.Nil {
			return nil
		}
		return RevealPick{
			Player:   chosen,
			Owner:    ctx.Controller(),
			Question: "Karn, Scion of Urza — choose one of the revealed cards; it goes into their hand and the other is exiled with a silver counter",
			Cards:    carry,
			Min:      1,
			Max:      1,
			Keyed:    s82OldB005CKarnPick,
		}.Apply(ctx)
	})

var s82OldB005CKarnPick = RegisterRevealPick("choose-cards/s82-old-b005-karn-scion-pick",
	func(ctx *Context, picked, left []uuid.UUID, _ []int) error {
		if err := ctx.Game.BounceCardsToHandKeyedForEffect(LiveCards(ctx.Game, picked), game.KeyedTail{}); err != nil {
			return err
		}
		for _, id := range LiveCards(ctx.Game, left) {
			if err := (ExileTarget{Target: id, ThenTail: ctx.Tail(s82OldB005CKarnSilverTail)}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

var s82OldB005CKarnSilverTail = CardTail("tail/s82-old-b005-karn-silver", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (AddCounter{Target: id, Kind: S82OldB005CSilverCounter, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB005CKarnScionMinusOne is the −1: "Put a card you own with a silver
// counter on it from exile into your hand." The controller picks among the
// cards they own in exile that carry a silver counter (any Karn's, since the
// counter is not linked to a source).
func S82OldB005CKarnScionMinusOne(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			if c.Owner == item.Controller && c.Counters[S82OldB005CSilverCounter] > 0 {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Karn, Scion of Urza — choose a card you own with a silver counter on it in exile to put into your hand",
		Cards:    ids,
		Min:      1,
		Max:      1,
		ThenKey:  s82OldB005CKarnTake,
	})
	return nil
}

var s82OldB005CKarnTake = game.RegisterChooseCardsThen("choose-cards/s82-old-b005-karn-scion-take", func(g *game.Game, r game.CardsChosen) error {
	var ids []uuid.UUID
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			ids = append(ids, id)
		}
	}
	return g.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
})
