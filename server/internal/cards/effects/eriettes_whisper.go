package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
			// The continuation is keyed (#4264), so the caster and the
			// creature (checked as the spell resolves, CR 608.2b) ride
			// the prompt as its Carry rather than the stack item.
			carry := []uuid.UUID{ctx.Controller()}
			if host, ok := ctx.ClauseTarget(1); ok && host.Kind == game.TargetCard {
				carry = append(carry, host.ID)
			}
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
				return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player: t.ID,
					Source: item.SourceCardID,
					N:      2,
				}, eriettesWhisperRole, carry)
			}
			return eriettesWhisperRoleOn(ctx.Game, item.SourceCardID, carry)
		},
	})
}

// eriettesWhisperRole is "create a Wicked Role token attached to up to
// one target creature you control", keyed (#4264): a table waiting on the
// discard is a restore point.
var eriettesWhisperRole = game.RegisterDiscardThen("choose-cards/eriettes-whisper-role", func(g *game.Game, r game.DiscardsDone) error {
	return eriettesWhisperRoleOn(g, r.Source, r.Carry)
})

// eriettesWhisperRoleOn is the Role: carry is the caster, then the
// creature when there is one.
func eriettesWhisperRoleOn(g *game.Game, source uuid.UUID, carry []uuid.UUID) error {
	if len(carry) < 2 {
		return nil
	}
	return CreateRoleToken{Role: RoleWicked, Host: carry[1]}.Apply(keyedContext(g, carry[0], source))
}
