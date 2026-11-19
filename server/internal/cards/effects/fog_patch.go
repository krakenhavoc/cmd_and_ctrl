package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fog Patch — Instant {1}{G} (Nemesis, #5839):
//
//	"Cast this spell only during the declare blockers step.
//	 Attacking creatures become blocked. (This spell works on creatures
//	 that can't be blocked.)"
//
// The timing is a CastCondition on the declare blockers step, on anyone's
// turn. Nobody has priority in that step until every defending player has
// declared blockers (#1501), so the spell always resolves after the
// declaration.
//
// "Attacking creatures" are the creatures attacking as the spell
// resolves, whoever controls them and whomever they attack. Each becomes
// a blocked creature with no creature blocking it (CR 509.1h,
// Game.BecomeBlockedForEffect) and stays blocked for the rest of combat.
// An attacker that was unblocked "becomes blocked", so its "whenever this
// creature becomes blocked" abilities trigger (CR 509.3c); one a creature
// already blocks does not become blocked again; and "becomes blocked by a
// creature" abilities do not trigger, because no creature blocks
// (CR 509.3d). A blocked creature with no creature blocking it assigns no
// combat damage (CR 510.1c) unless it has trample, which assigns all of
// it to the player, planeswalker or battle it is attacking (CR 702.19d).
// Evasion is not consulted: no block is declared, which is what the
// reminder text says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb8ff6c1-f9b4-4dae-8f46-0f1ad4c05355",
		Name:         "Fog Patch",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepDeclareBlockers
		},
		CastConditionLabel: "Cast this spell only during the declare blockers step.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ctx.Game.BecomeBlockedForEffect(AttackingCreatures(ctx.Game))
			return nil
		},
	})
}
