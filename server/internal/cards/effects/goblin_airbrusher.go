package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Airbrusher — Creature — Goblin Employee {1}{R}, 2/1:
//
//	"Whenever you place a sticker, create a Treasure token. If it's an art
//	 sticker, instead create two Treasure tokens."
//
// "Place a sticker" is putting one on an object (CR 123.3), announced as
// EventStickerPlaced with the sticker's kind (ADR 0208).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "906881da-6bb3-4a0e-9bd1-d1a5bc484144",
		Name:         "Goblin Airbrusher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouPlaceASticker("Goblin Airbrusher — create a Treasure token (two for an art sticker)",
				Do(CreateTokensForPlacedSticker{Template: TreasureToken(), N: 1, ArtN: 2})),
		},
	})
}
