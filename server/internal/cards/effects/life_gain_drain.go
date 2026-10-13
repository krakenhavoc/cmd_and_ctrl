package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// life_gain_drain.go — "Whenever you gain life, target opponent loses
// that much life." (Sanguine Bond, Vito, Thorn of the Dusk Rose).
//
// The trigger watches EventChangeLife for a positive delta on its
// controller, which every lifegain path emits (lifelink included), and
// captures the amount by value in Build, so a life total that changes
// again before the trigger resolves does not change the drain. The
// drain is life LOSS, not damage.
//
// Append-only: a new card printing the same sentence uses this row.

// WheneverYouGainLifeTargetOpponentLosesThatMuch is the row; label is
// the whole stack label.
func WheneverYouGainLifeTargetOpponentLosesThatMuch(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Target == source.Controller && ev.Amount > 0
		},
		Targets: TargetPlayer("target opponent", Opponent()),
		Key:     label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = ev.Amount
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Targets[0].ID, -item.Params.Amount)
		},
	}
}
