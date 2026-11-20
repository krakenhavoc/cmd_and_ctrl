package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lumberknot — Creature — Treefolk {2}{G}{G}, 1/1 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	"Hexproof
//	 Whenever a creature dies, put a +1/+1 counter on this creature."
//
// Blade of the Bloodchief's dies trigger (diedCreature reads the type the
// creature last had, CR 603.10a): any creature's death, tokens included,
// the Lumberknot's own too, which then finds no battlefield permanent to
// put the counter on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a9d6f32-c4a6-41ea-8188-235ea064dc53",
		Name:            "Lumberknot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died
			}, "Lumberknot — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
