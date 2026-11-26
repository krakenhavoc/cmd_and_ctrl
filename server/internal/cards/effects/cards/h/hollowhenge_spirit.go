package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hollowhenge Spirit — Creature — Spirit {3}{W}, 2/2 (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"Flash
//	 Flying
//	 When this creature enters, remove target attacking or blocking creature
//	 from combat."
//
// Flash and flying are the engine's keywords. The trigger is mandatory and
// targets any attacking or blocking creature (AttackingOrBlocking, judged as
// the trigger is put on the stack and again as it resolves); the removal is
// CR 506.4's, so the creature deals and receives no combat damage and stays
// on the battlefield, tapped or not as it was. With none to target the
// trigger is simply removed (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Hollowhenge Spirit — remove target attacking or blocking creature from combat"
	effects.Register(effects.Spec{
		OracleID:        "b34449d1-eab2-4be1-8351-568dc5ced7c3",
		Name:            "Hollowhenge Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := g.RemoveFromCombatForEffect(id); err != nil && err != game.ErrNotInCombat {
					return err
				}
				return nil
			}), effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking())),
		},
	})
}
