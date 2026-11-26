package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Totem-Guide Hartebeest — Creature — Antelope {4}{W}, 2/5 (Magic Origins):
//
//	"When this creature enters, you may search your library for an Aura
//	 card, reveal it, put it into your hand, then shuffle."
//
// Heliod's Pilgrim's search on a bigger body: "you may" is the prompt's
// decline, the filter is Aura cards, the pick is revealed and goes to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3a1bb65-556a-4016-a06a-3894710cb190",
		Name:         "Totem-Guide Hartebeest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Totem-Guide Hartebeest — search for an Aura card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.B36IsAuraCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Totem-Guide Hartebeest — an Aura card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
