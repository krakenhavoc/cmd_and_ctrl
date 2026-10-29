package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Getaway Barrel — Artifact {3}{R} (The Hobbit):
//
//	"When this artifact is put into a graveyard from the battlefield,
//	 reveal the top thirteen cards of your library. Put a random creature
//	 card from among them onto the battlefield. Put the rest on the
//	 bottom of your library in a random order."
//
// The reveal is public, the pick is not a choice: the creature cards among
// the thirteen are drawn from the game's keyed random stream
// (ChooseAtRandomForEffect, ADR 0054), so an undo or a restore picks the
// same card. The card enters through the ordinary library-to-battlefield
// batch (CR 614 entry replacements apply), and "the rest" is every other
// revealed card still in the library, on the bottom in a random order
// (PutRestOnBottomInRandomOrderTail), including the creature cards that were
// not picked. With no creature among them nothing enters and all thirteen
// go to the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ddbf891-7f80-44e9-b94c-70c1da0b91ed",
		Name:         "Getaway Barrel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Getaway Barrel — reveal the top thirteen cards; put a random creature card from among them onto the battlefield, the rest on the bottom in a random order",
				getawayBarrelTrigger),
		},
	})
}

func getawayBarrelTrigger(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 13,
		"Getaway Barrel — revealed from the top of the library")
	var creatures []uuid.UUID
	for _, id := range revealed {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	var picked uuid.UUID
	if len(creatures) > 0 {
		if got := g.ChooseAtRandomForEffect(randomDraw(ctx), creatures, 1); len(got) == 1 {
			picked = got[0]
		}
	}
	return PutFromLibraryOntoBattlefield{
		Player: item.Controller,
		Cards:  revealed,
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return picked != uuid.Nil && c.InstanceID == picked
		},
		All:      true,
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}
