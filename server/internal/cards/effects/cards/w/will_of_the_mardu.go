package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d8df9813-0376-4d30-8cdc-30451fb67726",
		Name:         "Will of the Mardu",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Create a number of 1/1 red Warrior creature tokens equal to the number of creatures target player controls.", effects.TargetPlayer("target player")),
			effects.Mode("Will of the Mardu deals damage to target creature equal to the number of creatures you control.", effects.TargetCreature("target creature")),
		).OrUpToIf(2, effects.YouControlACommander),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				for _, t := range effects.OptionTargets(ctx, 0) {
					n := effects.B14CreaturesControlled(ctx.Game, t.ID)
					if err := (effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 red Warrior"), N: n}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(1) {
				if err := ctx.Game.DamageInstanceForEffect(func() error {
					for _, t := range effects.OptionTargets(ctx, 1) {
						n := effects.B14CreaturesControlled(ctx.Game, ctx.Controller())
						if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
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
