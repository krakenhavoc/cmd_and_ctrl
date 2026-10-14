package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reconnaissance — Enchantment {W} (EDHREC top 1,000, #3654):
//
//	"{0}: Remove target attacking creature you control from combat and
//	 untap it. (If you activate during end of combat, the creature will
//	 untap after it deals combat damage.)"
//
// A free activated ability. The removal is the engine's CR 506.4
// (RemoveFromCombatForEffect), which leaves the creature on the
// battlefield with no combat damage dealt or taken; a creature that is
// no longer attacking when the ability resolves is an illegal target and
// is skipped (CR 608.2b). The untap follows the removal, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32599d7c-243b-447b-9d5d-e8e29830ac61",
		Name:         "Reconnaissance",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{0}: Remove target attacking creature you control from combat and untap it.",
			Cost:    ManaCost("{0}"),
			Targets: TargetCreature("target attacking creature you control", AttackingCreature(), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := g.RemoveFromCombatForEffect(id); err != nil {
					return nil
				}
				return UntapTarget{Target: id}.Apply(ctx)
			},
		}},
	})
}
