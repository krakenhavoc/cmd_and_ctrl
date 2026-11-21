package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "aa4f9618-8106-451e-889a-79a5e8159256",
		Name:         "Luxurious Libation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Tokens: 1, Targets: effects.ForTargets(effects.PumpXUntilEOT(0)).Targets},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := ctx.X()
			if err := (effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Luxurious Libation — +X/+X until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 1}.Apply(ctx)
		},
	})
}
