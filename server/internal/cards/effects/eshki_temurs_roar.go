package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eshki, Temur's Roar — Legendary Creature — Human Warrior {G}{U}{R}, 2/2
// (Tarkir: Dragonstorm Commander):
//
//	"Whenever you cast a creature spell, put a +1/+1 counter on Eshki. If
//	 that spell's power is 4 or greater, draw a card. If that spell's power
//	 is 6 or greater, Eshki deals damage equal to Eshki's power to each
//	 opponent."
//
// Each sentence is its own step of one trigger, so a prompt a step asks
// (a draw replacement) is waited out before the next one. The spell's power
// is read where the spell is now (PowerAnywhereForEffect: a countered one
// has gone to the graveyard but still has the power it was cast with). The
// damage reads Eshki's power AFTER the counter, from its last known value
// when it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b162dd3-3be6-406d-bf86-f7cc9eff098d",
		Name:         "Eshki, Temur's Roar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Creature(), "Eshki, Temur's Roar — +1/+1 counter, then draw and damage by the spell's power",
				Do(
					Instruction(plusOneCounterOnThisIfStillHere),
					Instruction(func(ctx *Context) error {
						if cmdrB13CastSpellPower(ctx) < 4 {
							return nil
						}
						return DrawCards{N: 1}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						if cmdrB13CastSpellPower(ctx) < 6 {
							return nil
						}
						n := sourcePowerNow(ctx)
						if n <= 0 {
							return nil
						}
						return damageToEachOpponent(ctx.Game, ctx.Item, n)
					}),
				)),
		},
	})
}
