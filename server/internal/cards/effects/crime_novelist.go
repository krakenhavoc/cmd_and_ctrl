package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crime Novelist — Creature — Goblin Bard {2}{R}, 1/3 (EDHREC rank
// 1357):
//
//	"Whenever you sacrifice an artifact, put a +1/+1 counter on this
//	 creature and add {R}."
//
// The Treasure deck's ritual engine: every Treasure cracked is a
// counter and a red mana back. The condition reads the sacrifice
// event, which fires while the permanent is still on the battlefield
// (Mirkwood Bats' shape), so the sacrificed card's type is intact.
// The mana is AddMana — this is a triggered ability that uses the
// stack, not a mana ability — so it lands in the pool when the
// trigger resolves and empties with the step (CR 106.4), which is
// why a Krark-Clan Ironworks chain nets an extra {R} per artifact
// only if the trigger is allowed to resolve first. The counter is
// skipped if the Novelist has left the battlefield by then; the mana
// still comes, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5996b1d0-7fe3-4fec-8730-9db901d887b1",
		Name:         "Crime Novelist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventSacrifice},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12YouSacrificedAnArtifact(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Crime Novelist — a +1/+1 counter, add {R}",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && z.Kind == game.ZoneBattlefield {
							if err := (AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return AddMana{Produced: "{R}"}.Apply(ctx)
					})
			},
		}},
	})
}
