package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assemble the Legion — Enchantment {3}{R}{W} (EDHREC rank 2212):
//
//	"At the beginning of your upkeep, put a muster counter on this
//	 enchantment. Then create a 1/1 red and white Soldier creature
//	 token with haste for each muster counter on this enchantment."
//
// The Boros inevitability engine: one Soldier the first upkeep, two
// the next, three after that. The upkeep trigger puts the counter
// on the Legion, then reads the count back OFF the Legion and makes
// that many Soldiers — so a Doubling Season sees the counter land
// and the Legion grows twice as fast, as printed. If the Legion has
// left the battlefield by the time the trigger resolves the counter
// cannot be placed, and the Soldiers are made for the muster
// counters it had when the trigger fired (CR 608.2h's last-known
// count, carried by the item since the LKI characteristic has no
// counters).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f81bef3-6ca0-4cf4-aa99-7b2813eeef04",
		Name:         "Assemble the Legion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				lastKnown := source.Counters["muster"]
				return game.NewTriggeredItem(source, "Assemble the Legion — add a muster counter, then muster the Soldiers",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						n := lastKnown
						if b15OnBattlefield(g, item.SourceCardID) {
							if err := (AddCounter{Target: item.SourceCardID, Kind: "muster", N: 1}).Apply(ctx); err != nil {
								return err
							}
							if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
								n = c.Counters["muster"]
							}
						}
						if n <= 0 {
							return nil
						}
						return CreateToken{Controller: item.Controller, Template: b20RedWhiteSoldierHasteToken(), N: n}.Apply(ctx)
					})
			},
		}},
	})
}
