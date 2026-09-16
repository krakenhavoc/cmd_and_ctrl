package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrion Feeder — 1/1 Creature — Zombie for {B}:
//
//	"This creature can't block.
//	 Sacrifice a creature: Put a +1/+1 counter on this creature."
//
// A free sac outlet that grows. Two things worth noting:
//
//   - The ability can eat the Feeder itself (CR 701.21 has no
//     "another" clause here). Doing so is legal and pointless: the
//     Feeder is already in the graveyard when the ability resolves,
//     so the counter has nowhere to land and the effect no-ops.
//   - "Can't block" is the drawback, and as of S24 the engine
//     enforces it. It is the first restriction a card prints on
//     ITSELF rather than handing to something it is attached to,
//     which is all RestrictSelf is: the same layer-6 write scoped by
//     selfOnly instead of by the attachment. CanBlock refuses the
//     pairing and internal/legal never offers it, so a bot seat is
//     not shown a block its own engine would bounce.
func init() {
	Register(Spec{
		OracleID:     "a1cc5e37-b09a-4b7f-afd5-77c1c35aa425",
		Name:         "Carrion Feeder",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice a creature: put a +1/+1 counter on this creature",
			Cost:  SacrificeACreature(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// The source may have been the sacrifice, or died in
				// response; AddCounter no-ops on a card that isn't
				// there.
				return AddCounter{
					Target: ctx.Source(),
					Kind:   "+1/+1",
					N:      1,
				}.Apply(ctx)
			},
		}},
	})
}
