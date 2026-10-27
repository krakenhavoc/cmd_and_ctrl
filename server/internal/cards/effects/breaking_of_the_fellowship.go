package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breaking of the Fellowship — Sorcery {1}{R}:
//
//	"Target creature an opponent controls deals damage equal to its
//	 power to another target creature that player controls. The Ring
//	 tempts you."
//
// The second target is related to the first (ControlledByTargetsController,
// #1943): it must be controlled by whoever controls the first, as the
// spell is cast (CR 601.2c) and again as it resolves (CR 608.2b). If the
// first creature has left the battlefield by then, "that player" is its
// last-known controller (CR 608.2h), so the second target can stay legal
// and the spell still resolves, and the Ring still tempts you.
//
// An illegal target neither deals nor is dealt damage (CR 608.2b), so
// the damage needs both targets; the tempt follows whenever the spell
// resolves, after any damage replacement has settled (ringTemptsYouNext).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96e081d2-ec72-41fc-ad81-2e31ab823ae0",
		Name:         "Breaking of the Fellowship",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature an opponent controls", OpponentControls()),
			Distinct(ControlledByTargetsController(TargetCreature("another target creature that player controls"), 0)),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			tempt := ringTemptsYouNext(item)
			biter, biterOK := ctx.ClauseTarget(0)
			victim, victimOK := ctx.ClauseTarget(1)
			if !biterOK || !victimOK {
				return tempt(ctx.Game)
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok || c.CurrentPower() <= 0 {
				return tempt(ctx.Game)
			}
			// The tempt is a keyed tail (#4253), the item riding it, so a
			// prompt the damage pauses on is a restore point.
			return ctx.Game.DealDamageEachKeyedForEffect(biter.ID, []uuid.UUID{victim.ID}, c.CurrentPower(),
				ctx.ItemTail(ringTemptsYouTail))
		},
	})
}
