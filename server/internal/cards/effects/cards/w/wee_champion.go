package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wee Champion — Creature — Minotaur Child Guest {R}, 0/1:
//
//	"Whenever you place a sticker, this creature gets +1/+1 until end of
//	 turn. If it's an art sticker, instead put a +1/+1 counter on this
//	 creature."
//
// "Place a sticker" is putting one on an object (CR 123.3), which the
// engine announces as EventStickerPlaced with the sticker's kind (ADR
// 0208); the kind decides the branch as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad8b90c0-4c66-4fc6-a749-cd10e6deabdb",
		Name:         "Wee Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouPlaceASticker("Wee Champion — it gets +1/+1 until end of turn (an art sticker: a +1/+1 counter instead)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if kind, _ := effects.PlacedStickerKind(item); kind == game.StickerArt {
						return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Wee Champion"}.Apply(ctx)
				}),
		},
	})
}
