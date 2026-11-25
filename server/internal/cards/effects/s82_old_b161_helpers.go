package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b161_helpers.go — shared bodies for the S82 older-sets batch 161
// (Modern Horizons, #5519).

// s82OldB161Shuffled is the pick answered: each chosen card still in the
// graveyard goes into its owner's library, and the library is shuffled if any
// moved (CR 701.24a). An on-disk key.
var s82OldB161Shuffled = game.RegisterChooseCardsThen("choose-cards/s82-old-b161-shuffle-graveyard-into-library", func(g *game.Game, r game.CardsChosen) error {
	moved := 0
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneLibrary); err != nil {
			return err
		}
		moved++
	}
	if moved == 0 {
		return nil
	}
	return g.ShuffleLibraryForEffect(r.Chooser)
})

// s82OldB161ShuffleUpToFour is the mill ended: "You shuffle up to four cards
// from your graveyard into your library", chosen now, from the graveyard as
// it stands after the mill (CR 608.2c).
var s82OldB161ShuffleUpToFour = game.RegisterTail("s82-old-b161/shuffle-up-to-four-from-graveyard", func(g *game.Game, o game.TailOutcome) error {
	cards := GyCardIDs(g, o.Tail.Player, nil)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  o.Tail.Player,
		Source:   o.Tail.Source,
		Question: "Stream of Thought — shuffle up to four cards from your graveyard into your library",
		Cards:    cards,
		Min:      0,
		Max:      4,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  s82OldB161Shuffled,
	})
	return nil
})

// S82OldB161MillThenShuffleUpToFour is "target player mills four cards. You
// shuffle up to four cards from your graveyard into your library": the mill
// runs through the replacement windows and the choice follows once it ends,
// whatever it moved.
func S82OldB161MillThenShuffleUpToFour(ctx *Context, target uuid.UUID) error {
	return ctx.Game.MillToZoneKeyedForEffect(target, 4, game.ZoneGraveyard, game.MillUntil{},
		game.KeyedTail{Ref: s82OldB161ShuffleUpToFour, Player: ctx.Controller(), Source: ctx.Source()})
}
