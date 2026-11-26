package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Surgery — Enchantment {1}{U} (New Phyrexia, S82 OLD batch 463,
// #5575):
//
//	"Whenever an opponent shuffles their library, you may look at the top
//	 two cards of that library. You may exile one of those cards. Then put
//	 the rest on top of that library in any order."
//
// The #6215 shuffle event (every shuffle counts, a search's included): the
// trigger's "you may" is asked as it triggers, the look is a look at
// another player's library (only the looker learns the cards, CR 401.2),
// the exile is optional, and the rest go back on top in an order the
// controller chooses (CR 401.4: the owner does not learn it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5128230-3535-4f77-8b97-20d0791e7b1d",
		Name:         "Psychic Surgery",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLibraryShuffled, effects.AnOpponentShufflesTheirLibrary,
				"Psychic Surgery — look at the top two cards of that library, you may exile one, then put the rest back on top in any order",
				psychicSurgeryLook),
				"Psychic Surgery — look at the top two cards of that library?"),
		},
	})
}

// psychicSurgeryLook looks at the top two cards of the shuffled library (the
// event's Actor owns it) and asks the controller which one, if any, to
// exile.
func psychicSurgeryLook(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	owner := item.Trigger.Event.Actor
	looked := g.LookAtTopOfPlayersLibraryForEffect(item.Controller, owner, 2)
	if len(looked) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: owner,
		Source:     item.SourceCardID,
		Question:   "Psychic Surgery — you may exile one of those cards; the rest go back on top in any order",
		Cards:      looked,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    psychicSurgeryExile,
		Carry:      looked,
	})
	return nil
}

// psychicSurgeryExile is the keyed answer: the card named is exiled, then
// whatever of the two is still in that library goes back on top in the
// controller's order. An on-disk key: never renamed, never reused.
var psychicSurgeryExile = game.RegisterChooseCardsThen("choose-cards/s82-old-b463-psychic-surgery-exile", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := g.ExileCardForEffect(id); err != nil {
			return err
		}
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if len(r.Picked) > 0 && id == r.Picked[0] {
			continue
		}
		rest = append(rest, id)
	}
	if len(rest) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   r.Chooser,
		Source:    r.Source,
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Reason:    "Psychic Surgery — put the rest on top of that library in any order",
	})
})
