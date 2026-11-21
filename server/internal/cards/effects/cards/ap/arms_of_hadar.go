package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arms of Hadar — Sorcery {3}{B} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Creatures target player controls get -2/-2 until end of turn."
//
// Gnashing of Teeth's second bullet at -2/-2: the targeted player's creatures
// as the spell resolves (CR 611.2c), locked then, so a creature that arrives
// later is not shrunk. With the player gone the spell is countered by the
// rules (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c15e5cb8-b94c-4157-9327-410e66606b82",
		Name:         "Arms of Hadar",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, OpponentsOnly: true}},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p, ok := ctx.ClauseTarget(0)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.ControlledByPlayer(p.ID)), Power: -2, Toughness: -2,
				Label: "Arms of Hadar — -2/-2",
			}.Apply(ctx)
		},
	})
}
