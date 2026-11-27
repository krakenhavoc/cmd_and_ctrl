package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b696_helpers.go — the Mercadian Masques enters and dies triggers
// (#5619, S82).

// mmq696SacrificeUnlessPaid is the decline branch of the Rishadan pirates'
// "sacrifices a permanent of their choice unless they pay": the opponent who
// would not pay sacrifices a permanent of their choice (CR 701.21a, a player
// with none sacrifices nothing). An on-disk key: never renamed, never reused.
var mmq696SacrificeUnlessPaid = PayUnlessElse("pay-unless/s82-old-b696-rishadan-sacrifice", func(ctx *Context, r game.PayUnlessAnswer) error {
	ctx.Game.PlayerSacrificesForEffect(ctx.Source(), r.Chooser, SacrificeSpec("a permanent", Permanent()),
		"Rishadan — sacrifice a permanent")
	return nil
})

// RishadanEachOpponentSacrificesUnlessTheyPay is the enters trigger of the
// Rishadan Cutpurse, Footpad and Brigand: "each opponent sacrifices a
// permanent of their choice unless they pay {N}". Each opponent still in the
// game is asked in turn order, and only one who does not pay sacrifices
// (Spelltithe Enforcer's pay-unless, aimed at every opponent). `cost` is the
// printed {N}.
func RishadanEachOpponentSacrificesUnlessTheyPay(name, cost string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, opp := range ctx.Opponents() {
			if err := (PayUnless{
				Chooser:  opp,
				Cost:     cost,
				Question: name + " — pay " + cost + "? (Otherwise sacrifice a permanent)",
				Else:     mmq696SacrificeUnlessPaid,
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
