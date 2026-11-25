package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b147_helpers.go — shared bodies for S82 OLD batch 147 (#5516).

// OldB147AnotherAttackerGetsXX is "another target attacking creature gets
// +X/+X until end of turn, where X is this creature's power" as a trigger
// body (Wildwood Mentor, Syr Faren): the target is chosen as the trigger goes
// on the stack, X is the source's power as it resolves or its last-known
// power if it has left (CR 608.2h), a negative power gives +0/+0 (CR 107.1b),
// and a target that is gone gets nothing (CR 608.2b).
func OldB147AnotherAttackerGetsXX(label string) game.TriggeredAbility {
	attack := WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		x := max(SourcePowerNow(ctx), 0)
		if x == 0 {
			return nil
		}
		return BoostUntilEOT{Target: target, Power: x, Toughness: x, Label: label}.Apply(ctx)
	})
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	return attack
}

// OldB147AttackerGainsUntilEndOfTurn is "whenever this creature attacks,
// another target attacking <kind> gains <keyword> until end of turn" as a
// trigger (Ogre Errant): the target clause carries the kind, and a target that
// is gone gains nothing (CR 608.2b).
func OldB147AttackerGainsUntilEndOfTurn(label, clause string, keyword string, kind CardPredicate) game.TriggeredAbility {
	attack := WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return GrantKeywordUntilEOT{Target: target, Keywords: []string{keyword}, Label: label}.Apply(ctx)
	})
	attack.Targets = Another(TargetCreature(clause, AttackingCreature(), kind))
	return attack
}
