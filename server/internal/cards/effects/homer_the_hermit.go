package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Homer, the Hermit — Legendary Creature — Crab Druid {B}{G}{U}, 0/9
// (Mystery Booster Commander Edition):
//
//	"Landfall — Whenever a land you control enters, any number of target
//	 players each mill X cards, where X is twice the number of Crabs,
//	 Lobsters, Nautiluses, Starfish, and/or Trilobites you control."
//
// The targets are any number of players (zero is legal), chosen as the
// trigger goes on the stack and re-checked as it resolves (CR 608.2b). X
// is counted once as the trigger resolves and every chosen player mills
// that many, in turn order. Homer is a Crab, so it counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a13aa603-8709-43aa-b69a-ee74de79561f",
		Name:         "Homer, the Hermit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Landfall("Homer, the Hermit — any number of target players each mill X cards, where X is twice the number of Crabs, Lobsters, Nautiluses, Starfish, and/or Trilobites you control",
					Do(Instruction(homerMill))),
				TargetPlayer("any number of target players").WithCount(0, 0)),
		},
	})
}

func homerMill(ctx *Context) error {
	n := 2 * countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
		for _, s := range []string{"Crab", "Lobster", "Nautilus", "Starfish", "Trilobite"} {
			if c.HasSubtype(s) {
				return true
			}
		}
		return false
	})
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (MillCards{Player: t.ID, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
