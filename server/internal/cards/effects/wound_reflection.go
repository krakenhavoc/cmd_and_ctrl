package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wound Reflection — Enchantment {5}{B} (EDHREC rank 1947):
//
//	"At the beginning of each end step, each opponent loses life
//	 equal to the life they lost this turn. (Damage causes loss of
//	 life.)"
//
// The six-mana damage doubler. "Each end step" is EventBeginEndStep
// on every player's turn, with no controller check; "the life they
// lost this turn" is read off the event log the way Bloodchief
// Ascension reads it — a negative EventChangeLife and an
// EventDealDamage to the player both count, summed back to the
// current turn's upkeep (b18LifeLostThisTurn). Every opponent's
// amount is computed before any is applied, so one opponent's loss
// to this trigger never inflates another's. Two Wound Reflections
// compound, as printed: the second sees the first's loss.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b09206a4-8b73-4125-8b79-53f6fd511b16",
		Name:         "Wound Reflection",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(_ game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return true
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Wound Reflection — each opponent loses the life they lost this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						opponents := ctx.Opponents()
						amounts := make([]int, len(opponents))
						for i, opp := range opponents {
							amounts[i] = b18LifeLostThisTurn(g, opp)
						}
						for i, opp := range opponents {
							if amounts[i] <= 0 {
								continue
							}
							if err := g.ChangePlayerLifeForEffect(item.SourceCardID, opp, -amounts[i]); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
