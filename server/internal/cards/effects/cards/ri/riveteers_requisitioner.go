package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riveteers Requisitioner — Creature — Lizard Rogue {1}{R}, 3/1:
//
//	"When this creature dies, create a Treasure token.
//	 Blitz {2}{R}"
//
// Blitz (CR 702.152, #3772): a blitzed Requisitioner sacrificed at the
// end step triggers both its own dies ability and the blitz draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "469f25de-c180-47e9-9b1b-e23269139bbd",
		Name:             "Riveteers Requisitioner",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Blitz("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Riveteers Requisitioner — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
