package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jadzi, Steward of Fate // Oracle's Gift — Legendary Creature — Human
// Wizard {2}{U}, 2/4, with the prepare spell Oracle's Gift, Sorcery
// {X}{X}{U} (Secrets of Strixhaven Commander, #4834):
//
//	"Jadzi enters prepared. (While it's prepared, you may cast a copy of
//	 its spell. Doing so unprepares it.)
//	 When Jadzi enters, draw two cards, then discard two cards."
//
//	Oracle's Gift: "Create X 0/0 green and blue Fractal creature tokens,
//	 then put X +1/+1 counters on each Fractal you control."
//
// A preparation card (CR 722): the permanent under the bare oracle ID and
// its prepare spell under "#1". "Enters prepared" is SelfEntersPrepared.
// The enters trigger draws both cards before the discard prompt opens, so
// a card just drawn is a legal discard (CR 608.2c).
//
// Oracle's Gift: the X tokens are created first and the counters are then
// PUT on every Fractal the caster controls as the second step runs: the
// ones just made and any Fractal already there, a nontoken one included,
// X on each. With X = 0 nothing is made and the older Fractals get no
// counters. The copy cast from exile chooses its own X as it is cast
// (CR 107.3a), and {X}{X} costs twice that in generic mana.
//
// No simplification.
func init() {
	const oracle = "c63b3511-d597-4c6e-9094-eabdeaa8a394"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Jadzi, Steward of Fate",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jadzi, Steward of Fate — draw two cards, then discard two cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 2, 2, "")
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Oracle's Gift",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.X() <= 0 {
					return nil
				}
				_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.TokenCard("0/0 green and blue Fractal")), ctx.X(), game.TokenEntryOptions{})
				return err
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				for _, id := range effects.SocB18EachFractalYouControl(ctx) {
					if !effects.OnBattlefield(ctx.Game, id) {
						continue
					}
					if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
