package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace's Phantasm — Creature — Illusion {U}, 1/1 (Magic 2013):
//
//	"Flying
//	 This creature gets +4/+4 as long as an opponent has ten or more cards in
//	 their graveyard."
//
// Swimmer in Nightmares' graveyard threshold, narrowed to one opponent's
// graveyard: a layer 7c modification read live, so it applies the moment any
// opponent's graveyard reaches ten cards and ends when it falls below.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab56df8f-0ba3-4d66-bde2-4c9b4acf07e5",
		Name:            "Jace's Phantasm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +4/+4 as long as an opponent has ten or more cards in their graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if target.InstanceID != source.InstanceID {
					return false
				}
				for _, p := range g.Seats {
					if p.ID != source.Controller && p.Graveyard != nil && p.Graveyard.Size() >= 10 {
						return true
					}
				}
				return false
			},
			Apply: effects.PumpBy(4, 4),
		}},
	})
}
