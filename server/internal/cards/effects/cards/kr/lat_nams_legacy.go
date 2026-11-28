package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lat-Nam's Legacy — Instant {1}{U} (Alliances, #5634):
//
//	"Shuffle a card from your hand into your library. If you do, draw two
//	 cards at the beginning of the next turn's upkeep."
//
// The controller picks a card from their hand (an empty hand does nothing),
// it goes into their library, the library is shuffled, and only then, if the
// card really reached the library, the two-card draw is scheduled for the next
// upkeep at the table (CR 603.7a). The pick, the tuck (which may pause for a
// commander) and the shuffle each wait on the one before.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f885ee54-3b26-4d61-9783-3930b0e625aa",
		Name:         "Lat-Nam's Legacy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			p := ctx.PlayerByID(ctx.Controller())
			if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
				return nil
			}
			ids := make([]uuid.UUID, 0, p.Hand.Size())
			for _, c := range p.Hand.Cards {
				ids = append(ids, c.InstanceID)
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   item.SourceCardID,
				Question: "Lat-Nam's Legacy — shuffle a card from your hand into your library",
				Cards:    ids,
				Min:      1,
				Max:      1,
				LiveZone: game.ZoneHand,
				ThenKey:  oldB814LatNamsPick,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// oldB814LatNamsPick tucks the chosen card; oldB814LatNamsTucked follows. An
// on-disk key.
var oldB814LatNamsPick = game.RegisterChooseCardsThen("choose-cards/old-b814-lat-nams-legacy-pick", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect(r.Picked, game.TuckOptions{}, game.KeyedTail{
		Ref: oldB814LatNamsTucked, Player: r.Chooser, Source: r.Source, Asked: game.NewAskedItem(r.Asked),
	})
})

// oldB814LatNamsTucked is "shuffle ... If you do, draw two cards at the
// beginning of the next turn's upkeep", told the card that reached the library.
var oldB814LatNamsTucked = effects.CardTail("then/old-b814-lat-nams-legacy-tucked", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	if err := ctx.Game.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	return effects.OldB814DrawAtTheNextUpkeep("Lat-Nam's Legacy — draw two cards", uuid.Nil, 2).Apply(ctx)
})
