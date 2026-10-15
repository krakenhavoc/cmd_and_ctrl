package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "469f25de-c180-47e9-9b1b-e23269139bbd",
		Name:             "Riveteers Requisitioner",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Riveteers Requisitioner — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
