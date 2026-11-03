package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     oracle,
		Name:         "Jadzi, Steward of Fate",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jadzi, Steward of Fate — draw two cards, then discard two cards",
				func(g *game.Game, item *game.StackItem) error {
					return drawThenDiscard(g, item.Controller, item.SourceCardID, 2, 2, "")
				}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Oracle's Gift",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if ctx.X() <= 0 {
					return nil
				}
				_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green and blue Fractal")), ctx.X(), game.TokenEntryOptions{})
				return err
			}),
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				for _, id := range socB18EachFractalYouControl(ctx) {
					if !onBattlefield(ctx.Game, id) {
						continue
					}
					if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
