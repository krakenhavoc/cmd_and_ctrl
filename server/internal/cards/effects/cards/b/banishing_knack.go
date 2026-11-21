package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banishing Knack — Instant {U} (Eventide, #5590):
//
//	"Until end of turn, target creature gains "{T}: Return target nonland
//	 permanent to its owner's hand.""
//
// Retraction Helix's grant (ADR 0093): the creature's controller may tap it
// to bounce a nonland permanent until end of turn.
//
// No simplification.
// banishingKnackBounce is the granted bundle's key. An on-disk identity:
// never renamed, never reused.
const banishingKnackBounce = "banishing-knack/bounce"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f23621e-eceb-402e-a8d3-8e2b734762b3",
		Name:         "Banishing Knack",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: banishingKnackBounce,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Return target nonland permanent to its owner's hand",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target nonland permanent", effects.Not(effects.Land())),
				Effect:  effects.RetractionHelixBounceTarget,
			}},
			Text: "{T}: Return target nonland permanent to its owner's hand.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{banishingKnackBounce},
				Label:  "Banishing Knack — until end of turn, it can bounce",
			}.Apply(ctx)
		},
	})
}
