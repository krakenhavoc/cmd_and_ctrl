package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Will of the Mardu — Instant {2}{W} (EDHREC rank 2269):
//
//	"Choose one. If you control a commander as you cast this spell,
//	 you may choose both instead.
//	 • Create a number of 1/1 red Warrior creature tokens equal to
//	   the number of creatures target player controls.
//	 • Will of the Mardu deals damage to target creature equal to
//	   the number of creatures you control."
//
// A token army sized to someone else's board, or a removal spell
// sized to your own, and with a commander out, both. Each mode
// carries its own target; the counts are read as the spell resolves
// (b14CreaturesControlled — the printed "target player" may be you,
// and the damage mode counts your creatures at that moment).
//
// SANDBOX GAP, weaker than printed — the Drown in Dreams posture:
// the "choose both" rider is not offered. Each mode carries its own
// target, and a modal spec whose maximum is above one may carry a
// target on at most one option (per-mode target slots are the open
// multi-target work), so the card is "choose one" whether or not a
// commander is on the board. Never stronger: the wider choice is
// simply absent.
func init() {
	Register(Spec{
		OracleID:     "d8df9813-0376-4d30-8cdc-30451fb67726",
		Name:         "Will of the Mardu",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Choosing both modes when you control a commander isn't implemented — you always choose one."},
		Modes: ChooseOne(
			Mode("Create a number of 1/1 red Warrior creature tokens equal to the number of creatures target player controls.", TargetPlayer("target player")),
			Mode("Will of the Mardu deals damage to target creature equal to the number of creatures you control.", TargetCreature("target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0]
			if ctx.HasMode(0) && target.Kind == game.TargetPlayer {
				n := b14CreaturesControlled(ctx.Game, target.ID)
				return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 red Warrior"), N: n}.Apply(ctx)
			}
			if ctx.HasMode(1) && target.Kind == game.TargetCard {
				n := b14CreaturesControlled(ctx.Game, ctx.Controller())
				return DealDamage{Source: ctx.Source(), Target: target.ID, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
