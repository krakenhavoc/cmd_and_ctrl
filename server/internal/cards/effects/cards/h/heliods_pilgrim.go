package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heliod's Pilgrim — Creature — Human Cleric {2}{W}, 1/2 (EDHREC rank
// 3776):
//
//	"When this creature enters, you may search your library for an
//	 Aura card, reveal it, put it into your hand, then shuffle."
//
// The Aura tutor on legs. An ETB with the S22 search chooser: "you
// may" is the prompt's decline, the filter is Aura cards on the
// printed type line, the pick is revealed and goes to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fe01c7d-f435-44c1-82f1-0ec2c47d4704",
		Name:         "Heliod's Pilgrim",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heliod's Pilgrim — search for an Aura card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.B36IsAuraCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Heliod's Pilgrim — an Aura card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
