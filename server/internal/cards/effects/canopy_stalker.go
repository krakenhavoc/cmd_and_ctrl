package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Canopy Stalker — Creature — Cat, {3}{G}, 4/2:
//
//	"This creature must be blocked if able.
//	 When this creature dies, you gain 1 life for each creature that
//	 died this turn."
//
// #1684: MustBeBlocked (#1597) plus a dies trigger reading the
// table-wide per-turn tally (TurnTally.CreaturesDied, the count
// Mahadi reads). The count is taken when the trigger RESOLVES, so it
// includes the Stalker itself and anything that died after it with
// the trigger still on the stack — "each creature that died this
// turn" is read on resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1062763-1fcf-4605-9253-daa1687e8173",
		Name:         "Canopy Stalker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{MustBeBlocked()},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, ThisDied, "Canopy Stalker — gain 1 life for each creature that died this turn", canopyStalkerGainLife),
		},
	})
}

func canopyStalkerGainLife(g *game.Game, item *game.StackItem) error {
	n := b11CreaturesDiedThisTurn(g)
	if n <= 0 {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
}
