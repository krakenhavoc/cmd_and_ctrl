package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divert Disaster — Instant {1}{U} (Edge of Eternities, #4461):
//
//	"Counter target spell unless its controller pays {2}. If they do, you
//	 create a Lander token. (It's an artifact with "{2}, {T}, Sacrifice
//	 this token: Search your library for a basic land card, put it onto
//	 the battlefield tapped, then shuffle.")"
//
// Assimilate Essence's shape: the counter-unless-pays prompt goes to the
// spell's controller, and "if they do" is the OnPaid branch, which runs
// only once the {2} was really paid, with the caster as "you" and makes
// the Lander. A payer who cannot fund it has declined: the spell is
// countered and no Lander is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76fdbc2c-d2b1-466f-9230-06e187bcc9e6",
		Name:         "Divert Disaster",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{2}",
				Question: "Divert Disaster — pay {2} or your spell is countered",
				Then:     divertDisasterPaid,
			}.Apply(ctx)
		},
	})
}

// divertDisasterPaid is "If they do, you create a Lander token", keyed: the
// spell rides the prompt as the asking item, so Controller() is the caster
// and not the payer. A key's branches replace the engine's own "counter
// that spell", so the decline is named here. An on-disk key.
var divertDisasterPaid = PayUnlessBranches("pay-unless/divert-disaster-lander",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Template: b16LanderToken(), N: 1}.Apply(ctx)
	},
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return ctx.Game.CounterIfStillOnStackForEffect(r.Guards)
	})
