package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shamanic Revelation — Sorcery {3}{G}{G} (EDHREC rank 368):
//
//	"Draw a card for each creature you control.
//	 Ferocious — You gain 4 life for each creature you control with
//	 power 4 or greater."
//
// The go-wide deck's refill. Both counts are taken once, as the
// spell resolves, over the creatures the caster controls: the draw
// is per creature, the lifegain per creature with power 4 or
// greater (CurrentPower — effective power plus counters, so a
// pumped or counter-laden 3/3 counts). Ferocious's "if you control a
// creature with power 4 or greater" gate is implied by a count of
// zero gaining zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1d171de-1c6d-4fb9-817a-9c689c709f3d",
		Name:         "Shamanic Revelation",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			creatures, big := 0, 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != ctx.Controller() || !c.IsCreature() {
					continue
				}
				creatures++
				if c.CurrentPower() >= 4 {
					big++
				}
			}
			if err := (DrawCards{Player: ctx.Controller(), N: creatures}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 4 * big}.Apply(ctx)
		},
	})
}
