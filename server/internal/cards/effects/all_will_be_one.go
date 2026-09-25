package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// All Will Be One — Enchantment {3}{R}{R} (EDHREC rank 1401):
//
//	"Whenever you put one or more counters on a permanent or player,
//	 this enchantment deals that much damage to target opponent,
//	 creature an opponent controls, or planeswalker an opponent
//	 controls."
//
// The counters deck's damage engine: every +1/+1 counter is a ping
// and a Cathars' Crusade trigger is a Pyroclasm. The condition reads
// the counter-placement event and works out, from the event log,
// both how many counters landed (the event carries only the new
// total) and whether the source's controller is the one who put them
// there (the player whose spell or ability was resolving) — see
// b12CountersPlacedByYou for the attribution rules and why every one
// of them errs toward NOT firing. One event is one permanent, so
// "one or more counters on a permanent" is one trigger per permanent
// touched, which is the printed batching. The target clause is the
// printed one: an opponent, or a creature or planeswalker one
// controls.
//
// Sandbox simplifications, declared:
//
//   - "or player": poison and energy counters on players emit no
//     event the trigger can watch, so counters put on a PLAYER never
//     fire it. Weaker than printed.
//   - Counters whose placer the log cannot attribute — a loyalty
//     cost or a Saga's lore counter on your own permanent while an
//     opponent's spell was the last thing to resolve — do not fire
//     it. Weaker, never stronger.
func init() {
	Register(Spec{
		OracleID:     "477374dc-042c-48f7-9ebe-99c15d8ae04f",
		Name:         "All Will Be One",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Counters put on a player (poison, energy) don't trigger it — only counters put on permanents do.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b12CountersPlacedByYou(ev, source, g)
				return ok
			},
			Targets: b12TargetOpponentOrTheirCreatureOrPlaneswalker(),
			Key:     "All Will Be One — that much damage",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount, _ := b12CountersPlacedByYou(ev, source, g)
				item := game.NewTriggeredItem(source, "All Will Be One — that much damage", nil)
				item.Params.Amount = amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
