package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noble Ox — Creature — Noble Ox {3}{W}, 2/4:
//
//	"Flash
//	 When Noble Ox enters the battlefield, all unblocked creatures
//	 attacking you become blocked by Noble Ox.
//	 Noble Ox can block any number of creatures."
//
// The trigger is game.BlockUnblockedAttackersForEffect (CR 509.1h, 509.3c):
// every attacker that is an unblocked creature at that time and attacks
// its controller (the player, not a planeswalker they control — "attacking
// you") becomes blocked by the Ox. It was not declared as a blocker, so no
// "whenever this blocks" ability fires. The block-count static is
// CanBlockAnyNumber, the same one Palace Guard reads.
func init() {
	Register(Spec{
		OracleID:        "b152753f-0516-4e21-b823-608e9d0c9c82",
		Name:            "Noble Ox",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{CanBlockAnyNumber(selfOnly)},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Noble Ox — all unblocked creatures attacking you become blocked by Noble Ox",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					g.BlockUnblockedAttackersForEffect(item.SourceCardID, item.Controller)
					return nil
				}),
		},
	})
}
