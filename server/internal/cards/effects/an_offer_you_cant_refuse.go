package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// An Offer You Can't Refuse — Instant for {U}:
//
//	"Counter target noncreature spell. Its controller creates two
//	 Treasure tokens."
//
// A one-mana counterspell that pays its victim off. In a Treasure
// deck the drawback is smaller than it looks and the tempo is the
// point.
//
// The two Treasures go to the COUNTERED spell's controller, not to
// you — the one detail worth getting right, and the reason this
// card resolves the target's controller rather than using
// item.Controller.
func init() {
	Register(Spec{
		OracleID: "234a734b-ba28-4f1b-9d01-3c3e7d516590",
		Name:     "An Offer You Can't Refuse",
		Targets:  TargetSpell("target noncreature spell", Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := controllerOfTarget(ctx, target)
			if err := (CounterTarget{StackID: target}).Apply(ctx); err != nil {
				return err
			}
			if !ok {
				return nil
			}
			return CreateToken{
				Controller: victim,
				Template:   TreasureToken(),
				N:          2,
			}.Apply(ctx)
		},
	})
}
