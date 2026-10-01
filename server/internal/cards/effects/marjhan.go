package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marjhan — Creature — Serpent {5}{U}{U}, 8/8:
//
//	"This creature doesn't untap during your untap step.
//	 {U}{U}, Sacrifice a creature: Untap this creature. Activate only
//	 during your upkeep.
//	 This creature can't attack unless defending player controls an
//	 Island.
//	 {U}{U}: This creature gets -1/-0 until end of turn and deals 1
//	 damage to target attacking creature without flying.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a);
// the sacrifice is §1's CR 603.8 state trigger. Both read one
// PermanentQuery. The rest are existing shapes:
//
//   - "Doesn't untap during your untap step" is the catalog's untap-step
//     restriction (CR 502.3).
//   - The untap row's sacrifice is a cost (CR 602.1a), so any creature
//     its controller controls pays it, Marjhan included, and "only during
//     your upkeep" is an activation condition (CR 602.1b).
//   - The ping shrinks Marjhan first and then deals the damage, as printed;
//     the target must be attacking and without flying, re-checked on
//     resolution (CR 608.2b).
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:              "3fdee2ab-7ec6-4fc6-ad99-f04571f94583",
		Name:                  "Marjhan",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Marjhan — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		Activated: []ActivatedAbility{
			{
				Label:     "{U}{U}, Sacrifice a creature: Untap this creature. Activate only during your upkeep.",
				Cost:      Plus(ManaCost("{U}{U}"), SacrificeACreature()),
				Condition: DuringYourUpkeep(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{U}{U}: This creature gets -1/-0 until end of turn and deals 1 damage to target attacking creature without flying.",
				Cost:    ManaCost("{U}{U}"),
				Targets: TargetCreature("target attacking creature without flying", AttackingCreature(), WithoutKeyword("flying")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if onBattlefield(g, item.SourceCardID) {
						if err := (BoostUntilEOT{Target: item.SourceCardID, Power: -1, Label: "Marjhan — -1/-0"}).Apply(ctx); err != nil {
							return err
						}
					}
					if len(item.Targets) == 0 {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 1}.Apply(ctx)
				},
			},
		},
	})
}
