package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bazaar Krovod — Creature — Beast {4}{W}, 2/5 (Return to Ravnica, S82 OLD):
//
//	"Whenever this creature attacks, another target attacking creature gets
//	 +0/+2 until end of turn. Untap that creature."
//
// Wildwood Mentor's "another target attacking creature" clause, chosen as the
// trigger goes on the stack. Both instructions read the same target: +0/+2,
// then untap (an untapped attacker stays attacking, CR 506.4). A target that
// has left combat or the battlefield gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Bazaar Krovod — another target attacking creature gets +0/+2 until end of turn. Untap that creature"
	attack := effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target := effects.FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		if err := (effects.BoostUntilEOT{Target: target, Toughness: 2, Label: "Bazaar Krovod — +0/+2 until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		return effects.UntapTarget{Target: target}.Apply(ctx)
	})
	attack.Targets = effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))
	effects.Register(effects.Spec{
		OracleID:     "1de561c7-4ff4-4eb9-8cbe-a29f568e786a",
		Name:         "Bazaar Krovod",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}
