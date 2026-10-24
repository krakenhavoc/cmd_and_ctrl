package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// upkeepSacrificeSource is the "or else" of "sacrifice <this permanent>
// unless you pay ...", keyed (#4264): the permanent is the prompt's
// source, so a table waiting on the payer is a restore point. Stasis,
// Haazda Shield Mate, Koskun Falls and every cumulative upkeep that is
// not a coin flip share it.
var upkeepSacrificeSource = PayUnlessElse("pay-unless/upkeep-sacrifice-source",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return SacrificePermanent{Target: r.Source}.Apply(ctx)
	})
