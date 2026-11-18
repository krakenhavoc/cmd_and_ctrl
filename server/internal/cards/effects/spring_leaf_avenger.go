package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spring-Leaf Avenger — Creature — Insect Ninja {3}{G}{G}, 6/5 (Neon Dynasty):
//
//	"Ninjutsu {3}{G}
//	 Whenever this creature deals combat damage to a player, return target
//	 permanent card from your graveyard to your hand."
//
// The target is chosen as the trigger goes on the stack, so a card put into
// the graveyard by the damage itself can be the one returned; it is
// re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19797dd6-6f48-47bc-a9ac-84d4acad3fa5",
		Name:         "Spring-Leaf Avenger",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsCombatDamageToAPlayer("Spring-Leaf Avenger — return target permanent card from your graveyard to your hand",
				returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent())),
		},
	})
}
