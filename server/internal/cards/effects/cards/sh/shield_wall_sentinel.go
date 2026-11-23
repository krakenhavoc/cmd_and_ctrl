package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shield-Wall Sentinel — Artifact Creature — Golem {4}, 1/3 (Dominaria
// United):
//
//	"Defender
//	 When this creature enters, you may search your library for a creature
//	 card with defender, reveal it, put it into your hand, then shuffle."
//
// Stoneforge Mystic's optional, revealed search. Your library is shuffled
// even if you find nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4976f08c-6f8e-42e4-9584-f04793518ddb",
		Name:            "Shield-Wall Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shield-Wall Sentinel — search for a creature card with defender", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && game.HasKeyword(&c, "defender")
					},
					Dest:     game.ZoneHand,
					Limit:    1,
					Reveal:   true,
					Shuffle:  true,
					Optional: true,
					Reason:   "Shield-Wall Sentinel — a creature card with defender, revealed, to hand",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
