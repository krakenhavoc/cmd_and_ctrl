package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodthirsty Conqueror — Creature — Vampire Knight {3}{B}{B}, 5/5
// (EDHREC rank 846):
//
//	"Flying, deathtouch
//	 Whenever an opponent loses life, you gain that much life."
//
// Exquisite Blood on a 5/5 flier — every point an opponent loses,
// from any source, is a point for you. "Loses life" is read off both
// of the engine's life-loss paths (b04OpponentLostLife, Exquisite
// Blood's helper): a drain
// emits a life-change event, but damage to a player — combat
// included — emits only the damage event and changes the total
// directly, and missing that would miss most of the game.
//
// The gain is the trigger's own life change, positive, so it never
// re-triggers itself. Paired with Sanguine Bond it loops, as in
// paper; that is the printed interaction, not a bug.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fd505c02-c59e-476d-8e88-35da862ddc23",
		Name:            "Bloodthirsty Conqueror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b04OpponentLostLife(ev, source.Controller, g)
				return ok
			},
			Key: "Bloodthirsty Conqueror — you gain that much life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount, _ := b04OpponentLostLife(ev, source.Controller, g)
				item := game.NewTriggeredItem(source, "Bloodthirsty Conqueror — you gain that much life", nil)
				item.Params.Amount = amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
