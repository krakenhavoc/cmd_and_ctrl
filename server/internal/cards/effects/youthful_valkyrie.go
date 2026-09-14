package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Youthful Valkyrie — Creature — Angel {1}{W}, 1/3 (EDHREC rank
// 3056):
//
//	"Flying
//	 Whenever another Angel you control enters, put a +1/+1 counter
//	 on this creature."
//
// The two-drop Angel that grows with the tribe. Flying rides
// PrintedKeywords; the trigger is Champion of the Perished's
// condition for Angels (b10AnotherPermanentWithSubtypeEnteredUnderYourControl
// — post-layer subtypes, so a token and a changeling both count),
// and the counter goes on the Valkyrie if it is still on the
// battlefield when the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d37ba4b-ff56-4eec-9dc2-2d7f357dc9c9",
		Name:            "Youthful Valkyrie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, "Angel")
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Youthful Valkyrie — put a +1/+1 counter on it",
					func(g *game.Game, item *game.StackItem) error {
						if !onBattlefield(g, item.SourceCardID) {
							return nil
						}
						return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
