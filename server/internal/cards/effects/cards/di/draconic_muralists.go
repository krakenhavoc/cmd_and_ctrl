package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Draconic Muralists — {3}{G} Creature — Dragon Bard 4/3
//
//	"When this creature dies, you may search your library for a Dragon card,
//	 reveal it, put it into your hand, then shuffle."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4dfad4ec-9695-4f3a-8269-4cc00c26204c",
		Name:         "Draconic Muralists",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Draconic Muralists — you may search your library for a Dragon card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Dragon") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Draconic Muralists — you may search your library for a Dragon card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
