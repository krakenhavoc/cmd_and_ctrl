package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snap — Instant {1}{U} (EDHREC rank 273):
//
//	"Return target creature to its owner's hand. Untap up to two
//	 lands."
//
// A free bounce spell: two mana in, two lands back. The bounce is
// Unsummon's; the refund runs after it, in printed order.
//
// "Untap up to two lands" prints no "target" and no "you control" —
// printed, it is a resolution-time choice among every land at the
// table. UntapUpToLands is that choice: a prompt over every tapped
// land at the table, any controller's, queued after the bounce.
func init() {
	Register(Spec{
		OracleID:     "ac914d98-221e-426c-8a50-342896b15f9e",
		Name:         "Snap",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BounceToHand{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return UntapUpToLands{N: 2, Question: "Snap — untap up to two lands"}.Apply(ctx)
		},
	})
}
