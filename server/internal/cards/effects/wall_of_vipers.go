package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Vipers — Creature — Snake Wall {2}{B}, 2/4:
//
//	"Defender
//	 {3}: Destroy this creature and target creature it's blocking. Any player may activate this ability."
//
// #1863: the target is described by what the source is doing, so it is
// BlockedBySource, judged at announce and again at resolution (CR
// 608.2b): a creature that is no longer blocked, because the Wall was
// removed from combat or the attacker left it, is an illegal target and
// the ability does nothing at all (CR 608.2b), the Wall included. The
// destruction is part of the effect, not a cost (#1793), so the ability
// can be answered and a Wall that has left is not destroyed again.
// Any player may activate it (ADR 0106 §1), and the attacker's
// controller usually does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fcfe370-ba33-438f-bd96-2ae560f59df9",
		Name:            "Wall of Vipers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:     "{3}: Destroy this creature and target creature it's blocking. Any player may activate this ability.",
			Cost:      game.AbilityCost{Mana: "{3}"},
			AnyPlayer: true,
			Targets:   BlockedBySource(TargetCreature("target creature it's blocking")),
			Effect:    destroyThisAndBlockedCreature,
		}},
	})
}

// destroyThisAndBlockedCreature is Wall of Vipers' effect: with no legal
// target the ability does nothing (CR 608.2b); otherwise the source
// (still the same object) and the target are destroyed.
func destroyThisAndBlockedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := destroyThisPermanent(false)(g, item); err != nil {
		return err
	}
	return DestroyTarget{Target: target}.Apply(ctx)
}
