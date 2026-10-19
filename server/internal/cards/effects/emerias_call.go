package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emeria's Call // Emeria, Shattered Skyclave — Modal DFC. Front:
// Sorcery {4}{W}{W}{W}:
//
//	"Create two 4/4 white Angel Warrior creature tokens with flying.
//	 Non-Angel creatures you control gain indestructible until your next
//	 turn."
//
// Back: Land, pay 3 life or enter tapped (mdfc_lands.go).
//
// Two steps (CR 608.2c): the tokens, which can stop on a CR 616 ordering
// prompt, then the grant, so the grant waits for the answer. The grant is
// a scoped effect whose set is pinned when it is registered (CR 611.2c):
// the creatures that are on the battlefield and not Angels at that moment
// keep indestructible until the start of the caster's next turn; one that
// arrives later does not get it. The two tokens are Angels, so they are
// outside the set whether or not they have arrived. No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ec2a242-9068-4ee2-8ac8-8341cc570f56",
		Name:         "Emeria's Call",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("4/4 white Angel Warrior with flying"), N: 2},
			Instruction(func(ctx *Context) error {
				me := ctx.Controller()
				return ScopedEffectFor{
					Match:    And(Creature(), YouControl(), Not(OfSubtype(ctx.TextSubtype("Angel")))),
					Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
					Duration: DurationUntilYourNextTurn(ctx, me),
					Label:    "Emeria's Call — non-Angel creatures you control gain indestructible until your next turn",
				}.Apply(ctx)
			}),
		),
	})
}
