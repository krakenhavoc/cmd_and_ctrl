package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spring Splasher — Creature — Frog Beast {1}{U}, 2/1 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever this creature attacks, target creature defending player
//	 controls gets -3/-0 until end of turn."
//
// Giant Space Amoeba's attack trigger without the token: the target is
// chosen off the Splasher's own attack (TargetOfDefendingPlayer), so
// attacking a planeswalker or battle names that permanent's controller
// or protector as the defending player. With no creature to target the
// trigger is removed (CR 603.3d), and a target gone by resolution takes
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	attack := WheneverThisAttacks("Spring Splasher — target creature defending player controls gets -3/-0",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: -3, Label: "Spring Splasher — -3/-0 until end of turn"}.Apply(ctx)
		})
	attack.TargetsFrom = TargetOfDefendingPlayer("creature", "target creature defending player controls", Creature())
	Register(Spec{
		OracleID:     "f1136a22-b202-4688-a44c-a9979c91261a",
		Name:         "Spring Splasher",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}
