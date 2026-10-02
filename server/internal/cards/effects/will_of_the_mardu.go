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
// The commander rider is the conditional mode count #1590 built:
// `OrUpToIf(2, YouControlACommander)`, read at announce (CR 601.2b)
// and fixed from then on. With both chosen the bullets run in PRINTED
// order (CR 608.2c) whatever order they were clicked in, and that is
// observable: the Warriors are made first, so the damage bullet
// counts them.
func init() {
	Register(Spec{
		OracleID:     "d8df9813-0376-4d30-8cdc-30451fb67726",
		Name:         "Will of the Mardu",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Create a number of 1/1 red Warrior creature tokens equal to the number of creatures target player controls.", TargetPlayer("target player")),
			Mode("Will of the Mardu deals damage to target creature equal to the number of creatures you control.", TargetCreature("target creature")),
		).OrUpToIf(2, YouControlACommander),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				for _, t := range OptionTargets(ctx, 0) {
					n := b14CreaturesControlled(ctx.Game, t.ID)
					if err := (CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 red Warrior"), N: n}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(1) {
				if err := ctx.Game.DamageInstanceForEffect(func() error {
					for _, t := range OptionTargets(ctx, 1) {
						n := b14CreaturesControlled(ctx.Game, ctx.Controller())
						if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
