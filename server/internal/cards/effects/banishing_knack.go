package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6f23621e-eceb-402e-a8d3-8e2b734762b3",
		Name:         "Banishing Knack",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: banishingKnackBounce,
			Activated: []ActivatedAbility{{
				Label:   "{T}: Return target nonland permanent to its owner's hand",
				Cost:    TapCost(),
				Targets: TargetPermanent("target nonland permanent", Not(Land())),
				Effect:  retractionHelixBounceTarget,
			}},
			Text: "{T}: Return target nonland permanent to its owner's hand.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{banishingKnackBounce},
				Label:  "Banishing Knack — until end of turn, it can bounce",
			}.Apply(ctx)
		},
	})
}
