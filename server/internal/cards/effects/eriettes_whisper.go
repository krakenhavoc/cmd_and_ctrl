package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eriette's Whisper — Sorcery {3}{B}:
//
//	"Target opponent discards two cards. Create a Wicked Role token
//	 attached to up to one target creature you control. (If you control
//	 another Role on it, put that one into the graveyard. Enchanted
//	 creature gets +1/+0. When this token is put into a graveyard, each
//	 opponent loses 1 life.)"
//
// Two target clauses read by slot, as Cut In does: slot 0 is the player
// who discards (their own choice of cards, CR 701.9a), slot 1 is the
// "up to one" creature that gets the Role. Each is checked on its own
// at resolution, so a creature that left does not stop the discard.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "27ca3201-a374-4a9d-b0ba-93b391f36844",
		Name:         "Eriette's Whisper",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target opponent", Opponent()),
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// The Role is the rest of the discard's sentence, so it rides
			// the discard's continuation (ADR 0013 §5y): the answer is
			// ignored, which is "create a Role" being printed AFTER it.
			role := func(g *game.Game, _ game.PromptedDiscards) error {
				return createRoleOnClauseTarget(NewContext(g, item), 1, RoleWicked)
			}
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
				return ctx.Game.PlayerDiscardsThenForEffect(game.DiscardPrompt{
					Player: t.ID,
					Source: item.SourceCardID,
					N:      2,
				}, role)
			}
			return role(ctx.Game, game.PromptedDiscards{})
		},
	})
}
