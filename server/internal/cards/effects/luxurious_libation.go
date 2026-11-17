package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luxurious Libation — Instant {X}{G} (Streets of New Capenna):
//
//	"Target creature gets +X/+X until end of turn. Create a 1/1 green and
//	 white Citizen creature token."
//
// X is the value announced as the spell was cast. A target that is illegal
// as the spell resolves makes the whole spell fizzle (CR 608.2b), so no
// Citizen is made then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa4f9618-8106-451e-889a-79a5e8159256",
		Name:         "Luxurious Libation",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Tokens: 1, Targets: ForTargets(PumpXUntilEOT(0)).Targets},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := ctx.X()
			if err := (BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Luxurious Libation — +X/+X until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1}.Apply(ctx)
		},
	})
}
