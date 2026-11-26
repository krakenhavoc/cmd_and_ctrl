package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boonweaver Giant — Creature — Giant Monk {6}{W}, 4/4 (Magic 2015, #5556):
//
//	"When this creature enters, you may search your graveyard, hand, and/or
//	 library for an Aura card and put it onto the battlefield attached to this
//	 creature. If you search your library this way, shuffle."
//
// Say Its Name's three-zone and/or search (SearchLibrary{AlsoGraveyard,
// AlsoHand}) with Light-Paws' enters-attached placement: the Aura enters
// attached to the Giant, and one that can't legally enchant it stays where it
// was (CR 303.4i). The library is shuffled only if it was searched. Nothing
// is searched once the Giant has left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79eed0c6-1853-4032-9d12-aebf5e15e78c",
		Name:         "Boonweaver Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.WhenThisEnters("Boonweaver Giant — search for an Aura to put onto the battlefield attached to it",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.OnBattlefield(g, item.SourceCardID) {
							return nil
						}
						return effects.SearchLibrary{
							Player:           item.Controller,
							Predicate:        func(c game.Card) bool { return c.IsAura() },
							Dest:             game.ZoneBattlefield,
							Limit:            1,
							Shuffle:          true,
							AlsoGraveyard:    true,
							AlsoHand:         true,
							Reason:           "Boonweaver Giant — an Aura card from your graveyard, hand or library",
							EntersAttachedTo: game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID},
						}.Apply(effects.NewContext(g, item))
					}),
				"Boonweaver Giant — search your graveyard, hand, and/or library for an Aura card?"),
		},
	})
}
