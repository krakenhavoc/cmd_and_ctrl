package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shield Sphere — Artifact Creature — Wall {0}, 0/6 (Alliances, #5634):
//
//	"Defender
//	 Whenever this creature blocks, put a -0/-1 counter on it."
//
// Defender is an engine keyword carried as printed data (CR 702.3b). The
// block trigger fires once per block declaration however many attackers
// the Sphere blocks, and the counter is a layer 7c change (CR 613.4c), so
// its sixth block leaves it at 0 toughness and it dies (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "df9ec9b0-be22-4136-923b-11a71ca4b6c2",
		Name:            "Shield Sphere",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Shield Sphere — put a -0/-1 counter on it", func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: "-0/-1", N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
