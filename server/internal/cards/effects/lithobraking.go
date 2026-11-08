package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lithobraking — Instant {2}{R} (Edge of Eternities, #4461):
//
//	"Create a Lander token. Then you may sacrifice an artifact. When you
//	 do, Lithobraking deals 2 damage to each creature. (A Lander token
//	 is an artifact with "{2}, {T}, Sacrifice this token: Search your
//	 library for a basic land card, put it onto the battlefield tapped,
//	 then shuffle.")"
//
// The Lander is Horizon Explorer's (b16LanderToken), made first, so it
// is one of the artifacts you may sacrifice. The sacrifice is asked once
// the token is there (CR 608.2c), over the artifacts you control then
// (maySacrificeOne). "When you do" is a reflexive trigger (CR 603.12,
// Cornered Crook's shape): it is created only once an artifact has really
// gone, goes on the stack after Lithobraking finishes resolving, and the
// table can respond before the damage. Its damage is one event to each
// creature (damageEachMatching). It can wipe the board, so the card
// declares a Sweep.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "118c4be1-2204-4133-a80e-990ef96c7e35",
		Name:         "Lithobraking",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OnResolve: ResolveSteps(
			CreateToken{Template: b16LanderToken(), N: 1},
			Instruction(func(ctx *Context) error {
				return maySacrificeOne(ctx.Game, ctx.Item, "Lithobraking — sacrifice an artifact to deal 2 damage to each creature?",
					Artifact(), false, lithobrakingSacrificed)
			}),
		),
	})
}

// lithobrakingSacrificed raises the reflexive trigger once an artifact has
// really gone, keyed. An on-disk key.
var lithobrakingSacrificed = RegisterSacrificedOne("prompt-run/lithobraking-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return WhenYouDo("Lithobraking — 2 damage to each creature", lithobrakingSweepBody).Apply(ctx)
})

// lithobrakingSweepBody is the reflexive trigger: 2 damage to each
// creature, from Lithobraking. An on-disk key.
var lithobrakingSweepBody = game.ReflexiveBody("lithobraking/damage-each-creature", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	return damageEachMatching(NewContext(g, item), Creature(), 2)
}, nil)
