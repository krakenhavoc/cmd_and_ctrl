package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// abnormalEnduranceReturn is the "When this creature dies, return it ..."
// bundle the spell hands its target until end of turn. An on-disk key: never
// renamed, never reused.
const abnormalEnduranceReturn = "abnormal-endurance/return"

// Abnormal Endurance — Instant {1}{B} (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Until end of turn, target creature gets +2/+0 and gains "When this
//	 creature dies, return it to the battlefield tapped under its owner's
//	 control.""
//
// Infuse with Vitality's duration grant of the dies trigger (CR 613.6, last-known
// information per CR 603.10a; the returned creature is a new object that the
// grant never named) with a +2/+0 pump instead of deathtouch and life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edbf1b87-2d1e-47e6-a04e-a2b1646af7d9",
		Name:         "Abnormal Endurance",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: abnormalEnduranceReturn,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Abnormal Endurance — return it to the battlefield tapped",
					effects.ReturnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 2, Label: "Abnormal Endurance — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{Target: target, Keys: []string{abnormalEnduranceReturn},
				Label: "Abnormal Endurance — until end of turn, it returns when it dies"}.Apply(ctx)
		},
	})
}
