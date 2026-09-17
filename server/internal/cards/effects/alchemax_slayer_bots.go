package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alchemax Slayer-Bots —
//
// "When this creature enters, tap target creature an opponent controls and put
// a stun counter on it. (If a permanent with a stun counter would become
// untapped, remove one from it instead.)"
func init() {
	Register(Spec{
		OracleID:     "cdc7356c-9621-4420-9fda-c8b91df0ec7c",
		Name:         "Alchemax Slayer-Bots",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Targeting(WhenThisEnters("Alchemax Slayer-Bots — tap and stun target creature", func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return (AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx)
		}), TargetCreature("target creature an opponent controls", OpponentControls()))},
	})
}
