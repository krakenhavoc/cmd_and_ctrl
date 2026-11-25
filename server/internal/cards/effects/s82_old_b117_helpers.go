package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b117_helpers.go — shared bodies for the S82 older-sets batch 117
// (Ikoria: Lair of Behemoths, #5512): the two Ikoria cards that return a card
// from your graveyard to your hand without targeting it (Mythos of Brokkos,
// Death's Oasis), the choice made as the clause resolves, after the mill or
// the search that may put the card there (CR 608.2c).

// s82OldB117Returned is the pick answered: each chosen card still in the
// graveyard goes to its owner's hand. An on-disk key.
var s82OldB117Returned = game.RegisterChooseCardsThen("choose-cards/s82-old-b117-return-to-hand", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneHand); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB117ChooseFromGraveyardToHand is "return up to <max> <cards> from your
// graveyard to your hand" with the cards chosen as the clause resolves: the
// candidates are read now, the player picks between min and max of them, and
// each picked card still in the graveyard goes to hand. With no candidate
// nothing is asked.
func S82OldB117ChooseFromGraveyardToHand(g *game.Game, player, source uuid.UUID, question string, match func(game.Card) bool, min, max int) error {
	cards := GyCardIDs(g, player, match)
	if len(cards) == 0 {
		return nil
	}
	if min > len(cards) {
		min = len(cards)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: question,
		Cards:    cards,
		Min:      min,
		Max:      max,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  s82OldB117Returned,
	})
	return nil
}

// S82OldB117ReturnUpToTwoPermanents is Mythos of Brokkos's second sentence:
// "Return up to two permanent cards from your graveyard to your hand."
func S82OldB117ReturnUpToTwoPermanents(g *game.Game, player, source uuid.UUID) error {
	return S82OldB117ChooseFromGraveyardToHand(g, player, source,
		"Mythos of Brokkos — return up to two permanent cards from your graveyard to your hand",
		func(c game.Card) bool { return c.IsPermanent() }, 0, 2)
}

// s82OldB117ReturnPermanents is the search ended: the return follows it.
var s82OldB117ReturnPermanents = game.RegisterTail("s82-old-b117/return-up-to-two-permanents", func(g *game.Game, o game.TailOutcome) error {
	return S82OldB117ReturnUpToTwoPermanents(g, o.Tail.Player, o.Tail.Source)
})

// S82OldB117ReturnPermanentsTail is the tail that runs Mythos of Brokkos's
// return once its search has settled.
func S82OldB117ReturnPermanentsTail(player, source uuid.UUID) game.KeyedTail {
	return game.KeyedTail{Ref: s82OldB117ReturnPermanents, Player: player, Source: source}
}

// s82OldB117ReturnLesserCreature is the mill ended: Death's Oasis returns a
// creature card with mana value lower than the carried number (N).
var s82OldB117ReturnLesserCreature = game.RegisterTail("s82-old-b117/return-lesser-creature", func(g *game.Game, o game.TailOutcome) error {
	ceiling := o.Tail.N
	return S82OldB117ChooseFromGraveyardToHand(g, o.Tail.Player, o.Tail.Source,
		"Death's Oasis — return a creature card with lesser mana value from your graveyard to your hand",
		func(c game.Card) bool { return c.IsCreature() && c.ManaValue() < ceiling }, 1, 1)
})

// S82OldB117MillThenReturnLesserCreature is "mill two cards. Then return a
// creature card with lesser mana value than the creature that died from your
// graveyard to your hand." The mill runs through the replacement windows and
// the return follows once it ends, whatever it moved.
func S82OldB117MillThenReturnLesserCreature(g *game.Game, player, source uuid.UUID, n, diedManaValue int) error {
	return g.MillToZoneKeyedForEffect(player, n, game.ZoneGraveyard, game.MillUntil{},
		game.KeyedTail{Ref: s82OldB117ReturnLesserCreature, Player: player, Source: source, N: diedManaValue})
}
