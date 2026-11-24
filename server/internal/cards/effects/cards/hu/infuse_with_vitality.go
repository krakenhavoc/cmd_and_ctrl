package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// infuseWithVitalityReturn is the "When this creature dies, return it ..."
// bundle the spell hands its target until end of turn. An on-disk key: never
// renamed, never reused.
const infuseWithVitalityReturn = "infuse-with-vitality/return"

// Infuse with Vitality — {B}{G} Instant (Strixhaven: School of Mages, #5502):
//
//	"Until end of turn, target creature gains deathtouch and "When this
//	 creature dies, return it to the battlefield tapped under its owner's
//	 control."
//	 You gain 2 life."
//
// Feign Death's duration grant without the counter: the dies trigger is the
// creature's own ability until end of turn (CR 613.6), it fires from
// last-known information (CR 603.10a), and the returned creature is a new
// object (CR 400.7) that the grant never named, so it does not return a
// second time. Deathtouch is a keyword until end of turn on the same target.
// With its one target illegal on resolution the spell does nothing at all,
// the life included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6799217a-030c-4c1b-9de3-99725c6ef4cf",
		Name:         "Infuse with Vitality",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: infuseWithVitalityReturn,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Infuse with Vitality — return it to the battlefield tapped",
					effects.ReturnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"deathtouch"},
				Label: "Infuse with Vitality — deathtouch"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantAbilitiesFor{Target: target, Keys: []string{infuseWithVitalityReturn},
				Label: "Infuse with Vitality — until end of turn, it returns when it dies"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
