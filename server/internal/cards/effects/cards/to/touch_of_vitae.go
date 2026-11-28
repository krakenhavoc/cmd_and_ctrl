package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Touch of Vitae — Instant {2}{G} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Until end of turn, target creature gains haste and "{0}: Untap this
//	 creature. Activate only once."
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Banishing Knack's grant of an activated ability until end of turn (ADR 0093)
// plus a haste grant. "Activate only once" is the OnlyOnce bit (#5221, CR 602.5b),
// kept per object, so the creature may untap itself once. The delayed draw
// follows (CR 603.7a, next_upkeep_draw.go). With its one target gone the spell
// does not resolve, so nothing is drawn (CR 608.2b).
//
// No simplification.

// touchOfVitaeUntap is the granted bundle's key. An on-disk identity: never
// renamed, never reused.
const touchOfVitaeUntap = "touch-of-vitae/untap"

func init() {
	const label = "{0}: Untap this creature. Activate only once."
	effects.Register(effects.Spec{
		OracleID:     "84f1e596-f56f-48c9-9bff-2c0bb2cea2d3",
		Name:         "Touch of Vitae",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
		Grants: []effects.AbilityGrant{{
			Key: touchOfVitaeUntap,
			Activated: []effects.ActivatedAbility{{
				Label:    label,
				Cost:     effects.ManaCost("{0}"),
				Purpose:  game.Purpose{Answers: game.AnswerValue},
				OnlyOnce: true,
				Effect:   effects.UntapThisPermanentEffect,
			}},
			Text: "{0}: Untap this creature. Activate only once.",
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"}, Label: "Touch of Vitae — haste"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantAbilitiesFor{Target: target, Keys: []string{touchOfVitaeUntap}, Label: "Touch of Vitae — until end of turn, it can untap itself"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Touch of Vitae — draw a card").Apply(ctx)
		},
	})
}
