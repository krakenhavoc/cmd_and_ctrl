package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nav Squad Commandos — Creature — Human Soldier {4}{W}, 3/5 (Gatecrash, #5565):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature gets +1/+1 until end of turn. Untap it."
//
// Battalion is Boros Strike-Captain's group test; the pump comes first,
// then the untap, in printed order. The creature stays attacking (CR
// 506.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "affdf41e-c851-4f5d-bd3a-7ca7ada4416a",
		Name:         "Nav Squad Commandos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			gtcBattalion("Nav Squad Commandos — it gets +1/+1 until end of turn, then untap it (battalion)",
				func(g *game.Game, item *game.StackItem) error {
					if err := thisGetsUntilEndOfTurn(1, 1, "Nav Squad Commandos — +1/+1")(g, item); err != nil {
						return err
					}
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}
