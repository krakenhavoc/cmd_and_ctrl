package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Warlord — Legendary Enchantment {2}{R} (Reality Fracture,
// tracker #2795):
//
//	"When Way of the Warlord enters, empower Jace 5.
//	 Planeswalkers you control have "[−4]: This planeswalker deals 2
//	 damage to up to one target creature or planeswalker and 2 damage to
//	 target player.""
//
// Empower Jace is the keyword action (ADR 0139); the granted row is an
// ADR 0093 bundle with a loyalty cost (ADR 0140). Two target clauses,
// one damage instruction (Insult // Injury's shape): the two hits are
// one instance and each clause is rechecked on its own (CR 608.2b), so
// the player is hit even when the first clause has no target. The
// damage is dealt by the planeswalker.
//
// No simplification.
func init() {
	const grant = "way-of-the-warlord/damage"
	Register(Spec{
		OracleID:     "10721d62-dc68-43b4-bfa2-dcb8cf8ea980",
		Name:         "Way of the Warlord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Warlord — empower Jace 5", Do(EmpowerJace{N: 5})),
		},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label: "−4: This planeswalker deals 2 damage to up to one target creature or planeswalker and 2 damage to target player.",
				Cost:  LoyaltyCost(-4),
				Targets: Clauses(
					TargetPermanent("up to one target creature or planeswalker", Or(Creature(), Planeswalker())).WithCount(0, 1),
					TargetPlayer("target player")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.DamageInstanceForEffect(func() error {
						for slot := 0; slot < 2; slot++ {
							t, ok := ctx.ClauseTarget(slot)
							if !ok {
								continue
							}
							if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				},
			}},
			Text: "[−4]: This planeswalker deals 2 damage to up to one target creature or planeswalker and 2 damage to target player.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
