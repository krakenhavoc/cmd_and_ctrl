package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang and Katara — Legendary Creature — Human Avatar Ally {3}{G}{W}{U},
// 5/5 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever Aang and Katara enter or attack, create X 1/1 white Ally
//	 creature tokens, where X is the number of tapped artifacts and/or
//	 creatures you control."
//
// Sun Titan's one ability with two conditions. X is counted as the
// trigger resolves: each tapped permanent you control that is an
// artifact, a creature or both counts once, so an attacking Aang and
// Katara (tapped by the attack) counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "481c3e14-b670-4fab-aa9f-6ce5b514096d",
		Name:         "Aang and Katara",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Aang and Katara — create X 1/1 white Ally creature tokens",
				Do(Instruction(func(ctx *Context) error {
					n := countControlled(ctx.Game, ctx.Controller(), avB27TappedArtifactOrCreature)
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 white Ally"), N: n}.Apply(ctx)
				}))),
		},
	})
}
