package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Turret — Artifact {2}{R} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"At the beginning of combat on your turn, up to one target creature
//	 gets +2/+0 and gains firebending 2 until end of turn.
//	 {R}: Put a charge counter on this artifact.
//	 Remove fifty charge counters from this artifact: It deals 50 damage
//	 to any target."
//
// The granted firebending is Fire Nation Palace's bundle with N = 2: an
// attack trigger the creature has until end of turn, so it fires when
// that creature attacks this combat (the mana lasts until end of
// combat, the rulings). The fifty counters are a real cost
// (RemoveCountersFromThis), so the last ability can't be activated with
// fewer.
//
// No simplification.
func init() {
	const label = "Fire Nation Turret — up to one target creature gets +2/+0 and gains firebending 2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "2ce2a091-95b6-45e5-9097-77fea411162e",
		Name:         "Fire Nation Turret",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: fireNationTurretFirebending,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Firebending 2", effects.Do(effects.NewFirebending(2))),
			},
			Text: "Firebending 2 (Whenever this creature attacks, add {R}{R}. This mana lasts until end of combat.)",
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Fire Nation Turret — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantAbilitiesFor{Target: id, Keys: []string{fireNationTurretFirebending}, Label: "Fire Nation Turret — firebending 2 until end of turn"}.Apply(ctx)
				}),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "firebending"))),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: Put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.PutChargeCounterOnThis,
			},
			{
				Label:   "Remove fifty charge counters from this artifact: It deals 50 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 50)),
				Cost:    effects.RemoveCountersFromThis(game.CounterCharge, 50),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(50),
			},
		},
	})
}

// fireNationTurretFirebending is the granted bundle's key: an on-disk
// identity, never renamed.
const fireNationTurretFirebending = "fire-nation-turret/firebending-2"
