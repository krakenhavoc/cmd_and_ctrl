package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Will of the Temur — Sorcery {5}{U} (Tarkir: Dragonstorm Commander):
//
//	"Choose one. If you control a commander as you cast this spell, you may
//	 choose both instead.
//	 • Create a token that's a copy of target permanent, except it's a 4/4
//	   Dragon creature with flying in addition to its other types.
//	 • Target player draws cards equal to the greatest mana value among
//	   permanents you control."
//
// The mode count is Will of the Abzan's (`OrUpToIf(2, YouControlACommander)`,
// read at announce and fixed, the rulings), each bullet with its own target
// group, and the body runs the bullets in printed order (CR 608.2c). The copy
// takes the target's copiable values (Applied Geometry's door) and is then
// a Dragon creature in addition to its other types with base 4/4 and flying
// (CR 707.9a): the exception edits the token's own characteristics, so a copy
// of a land is a 4/4 land creature. A target that left leaves no token. The
// draw counts permanents as the second bullet resolves, so a token the first
// bullet just made, a copy of a six-drop, counts (X in a cost is 0).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebe1448e-21c5-4b73-8296-03d3e5097c50",
		Name:         "Will of the Temur",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Create a token that's a copy of target permanent, except it's a 4/4 Dragon creature with flying in addition to its other types.",
				TargetPermanent("target permanent")),
			Mode("Target player draws cards equal to the greatest mana value among permanents you control.",
				TargetPlayer("target player")),
		).OrUpToIf(2, YouControlACommander),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				for _, t := range OptionTargets(ctx, 0) {
					if t.Kind != game.TargetCard {
						continue
					}
					dragon := ctx.TextSubtype("Dragon")
					return CreateTokenCopy{
						Controller: ctx.Controller(),
						Copy:       t.ID,
						N:          1,
						Except: func(tok *game.Card) {
							tok.TypeLine = socB19WithCreatureAndSubtypeInAddition(tok.TypeLine, dragon)
							tok.Power, tok.Toughness = 4, 4
							tok.VariableToughness = false
							tok.PrintedPTKnown = true
							if !slices.Contains(tok.Keywords, "flying") {
								tok.Keywords = append(tok.Keywords, "flying")
							}
						},
					}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				for _, t := range OptionTargets(ctx, 1) {
					if t.Kind != game.TargetPlayer {
						continue
					}
					n := cmdrB13GreatestManaValueYouControl(ctx.Game, ctx.Controller())
					if n <= 0 {
						return nil
					}
					return DrawCards{Player: t.ID, N: n}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
