package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caves of Chaos Adventurer — Creature — Human Barbarian {3}{R}, 5/3 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"Trample
//	 When this creature enters, you take the initiative.
//	 Whenever this creature attacks, exile the top card of your library. If you've
//	 completed a dungeon, you may play that card this turn without paying its mana
//	 cost. Otherwise, you may play that card this turn."
//
// Trample is printed and the entry trigger is the shared take-the-initiative one
// (CR 725). The attack trigger is the impulse exile (ExileTopWithPermission) whose
// "if you've completed a dungeon" is read as the trigger resolves, after the card is
// exiled, to choose between the free permission and the ordinary one; a land can be
// played either way (it has no mana cost to skip). Both permissions last until
// the end of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4d2dcbb-9910-482e-9e46-dfed2097509b",
		Name:            "Caves of Chaos Adventurer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Caves of Chaos Adventurer"),
			effects.WheneverThisAttacks("Caves of Chaos Adventurer — exile the top card of your library; you may play it this turn, without paying its mana cost if you've completed a dungeon",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{
						From: item.Controller, GrantTo: item.Controller, N: 1,
						Free: effects.YouCompletedADungeon(g, item.Controller),
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
