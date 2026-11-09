package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b15_helpers.go — shared bodies for Standard set batch 15 (#4452).

// stdB15LookKeepOneOnTopRestToGraveyard is "look at the top three cards of
// your library. You may put one of those cards back on top of your
// library. Put the rest into your graveyard" (Gutless Plunderer). A look,
// so only the looker sees the cards; the pick is optional, and every card
// not picked is put (not milled) into the graveyard.
func stdB15LookKeepOneOnTopRestToGraveyard(ctx *Context, n int, label string) error {
	player := ctx.Controller()
	looked := ctx.Game.LookAtTopOfLibraryForEffect(player, n)
	if len(looked) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: label,
		Cards:    looked,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  stdB15KeepOneOnTop,
	})
	return nil
}

// stdB15KeepOneOnTop leaves the picked card where it is, on top of the
// library, and puts every other card that was looked at into the
// graveyard. An on-disk key.
var stdB15KeepOneOnTop = game.RegisterChooseCardsThen("choose-cards/std-b15-keep-one-on-top-rest-graveyard",
	func(g *game.Game, r game.CardsChosen) error {
		kept := map[uuid.UUID]bool{}
		for _, id := range r.Picked {
			kept[id] = true
		}
		var rest []uuid.UUID
		for _, id := range r.Candidates {
			if !kept[id] {
				rest = append(rest, id)
			}
		}
		return restIntoGraveyard(g, rest)
	})

// stdB15ExileTopTwoChooseOnePlay is "exile the top two cards of your
// library. Choose one of them. Until the end of your next turn, you may
// play that card" (Strongbox Raider): both cards are exiled face up, the
// controller picks one, and only it gets the window; the other stays
// exiled with nothing.
func stdB15ExileTopTwoChooseOnePlay(ctx *Context) error {
	ids := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), 2)
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(stdB15ExileChooseOne))
}

// stdB15ExileChooseOne is "choose one of them" over the cards that
// reached exile. An on-disk key.
var stdB15ExileChooseOne = CardTail("then/std-b15-strongbox-raider-choose", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Strongbox Raider — choose one of them (you may play it until the end of your next turn)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  stdB15ExilePlayGrant,
	})
	return nil
})

// stdB15ExilePlayGrant gives the chosen card its window. An on-disk key.
var stdB15ExilePlayGrant = game.RegisterChooseCardsThen("choose-cards/std-b15-strongbox-raider-play", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfYourNextTurnDuration(r.Chooser),
			Label:    "Strongbox Raider — play it until the end of your next turn",
		})
	}
	return nil
})

// stdB15ReturnDiedWithCounters is "return that card to the battlefield
// under its owner's control with <counters> on it" for the card an
// EventLTB named (Infernal Vessel, Valkyrie's Call). Nothing happens when
// the card is no longer in a graveyard (CR 608.2b); a token has no card to
// return. The counters are entry counters (CR 614.1c), and `tail` runs
// with the new object once it has landed.
func stdB15ReturnDiedWithCounters(g *game.Game, item *game.StackItem, counters map[string]int, tail game.TailRef) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	dead, ok := g.LookupCardForEffect(id)
	if !ok || dead.IsToken() {
		return nil
	}
	ctx := NewContext(g, item)
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, dead.Owner, false, counters, ctx.ItemTail(tail))
}

// stdB15LoyaltyCounterOnThis is "put a loyalty counter on <this planeswalker>"
// as a trigger's body (Kaito, Cunning Infiltrator): only while it is still on
// the battlefield.
func stdB15LoyaltyCounterOnThis(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterLoyalty, N: 1}.Apply(NewContext(g, item))
}
