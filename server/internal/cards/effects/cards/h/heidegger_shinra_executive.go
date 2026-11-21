package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heidegger, Shinra Executive — Legendary Creature — Human Soldier
// {3}{W}, 3/3 (#3656):
//
//	"At the beginning of combat on your turn, target creature you
//	 control gets +X/+0 until end of turn, where X is the number of
//	 Soldiers you control.
//	 At the beginning of your end step, create a number of 1/1 white
//	 Soldier creature tokens equal to the number of opponents who
//	 control more creatures than you."
//
// Reckless Stormseeker's targeted beginning-of-combat boost, with X
// counted as the ability resolves (Heidegger himself is a Soldier). The
// end-step trigger counts opponents, not creatures, and compares each
// opponent's creature count with yours when it resolves.
//
// No simplification.
func init() {
	const name = "Heidegger, Shinra Executive"
	effects.Register(effects.Spec{
		OracleID:     "215024ba-811e-4011-bbc8-8dd14a34f973",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat(name+" — target creature you control gets +X/+0 until end of turn, X the number of Soldiers you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						target, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						x := 0
						// "Soldiers" is a word of the source's text (CR 612.2, #3984).
						soldier := ctx.TextSubtype("Soldier")
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(soldier) {
								x++
							}
						}
						return effects.BoostUntilEOT{Target: target, Power: x, Label: name + " — +X/+0"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
			effects.AtYourEndStep(name+" — create a 1/1 white Soldier for each opponent who controls more creatures than you",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					mine, theirs := 0, map[uuid.UUID]int{}
					for _, c := range g.BattlefieldCardsForEffect() {
						if !c.IsCreature() {
							continue
						}
						if c.Controller == item.Controller {
							mine++
						} else {
							theirs[c.Controller]++
						}
					}
					n := 0
					for _, opp := range ctx.Opponents() {
						if theirs[opp] > mine {
							n++
						}
					}
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Soldier"), N: n}.Apply(ctx)
				}),
		},
	})
}
