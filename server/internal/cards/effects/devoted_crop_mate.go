package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devoted Crop-Mate — Creature — Human Warrior {2}{W}, 3/2:
//
//	"You may exert this creature as it attacks. When you do, return
//	 target creature card with mana value 2 or less from your graveyard
//	 to the battlefield. (An exerted creature won't untap during your
//	 next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a targeted linked
// trigger (CR 607.2h). The target is a creature card in YOUR graveyard
// with mana value 2 or less, re-checked as the trigger resolves (CR
// 608.2b). It returns untapped and not attacking (CR 508.4), under its
// owner's control, which is yours. With no such card the trigger is
// removed and the Crop-Mate stays exerted (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Devoted Crop-Mate — return target creature card with mana value 2 or less from your graveyard to the battlefield"
	Register(Spec{
		OracleID:      "fcde4643-a212-4a9e-94c3-0a0a95841bc6",
		Name:          "Devoted Crop-Mate",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenExerted(label, func(g *game.Game, item *game.StackItem) error {
					reanimateSingleTarget(NewContext(g, item), item.Controller)
					return nil
				}),
				TargetCardInGraveyard("target creature card with mana value 2 or less in your graveyard", Creature(), YouOwn(), ManaValueLE(2))),
		},
	})
}
