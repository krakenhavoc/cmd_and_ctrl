package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pride Guardian — Creature — Cat Monk {W}, 0/3 (Magic 2012, #5573):
//
//	"Defender
//	 Whenever this creature blocks, you gain 3 life."
//
// Defender, and Giant Badger's block trigger (once however many attackers
// it blocks, CR 509.3a) with a life gain for its body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6b2d396f-802d-4b49-8418-5a96b8a12226",
		Name:            "Pride Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Pride Guardian — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
