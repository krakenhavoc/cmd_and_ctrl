package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retraction Helix — Instant {U}:
//
//	"Until end of turn, target creature gains "{T}: Return target
//	 nonland permanent to its owner's hand.""
//
// A duration grant (ADR 0093 PR 4, #1584): the granted ability is the
// CREATURE's (ADR 0093 Decision 4). Its controller activates it, {T}
// taps it, and CR 302.6 refuses it on a creature that came under its
// controller's control this turn — so casting this on a creature you
// just cast does nothing until it has haste. It shows in the
// creature's ability menu labelled "from Retraction Helix", and it is
// gone at the cleanup step.
//
// No simplification.
const retractionHelixBounce = "retraction-helix/bounce"

func init() {
	Register(Spec{
		OracleID:     "2707f9f4-2b80-47ac-af0a-99fc53af94bf",
		Name:         "Retraction Helix",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: retractionHelixBounce,
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
				Keys:   []string{retractionHelixBounce},
				Label:  "Retraction Helix — until end of turn, it can bounce",
			}.Apply(ctx)
		},
	})
}

// retractionHelixBounceTarget is the granted ability's effect: return
// its target, if still legal, to its owner's hand.
func retractionHelixBounceTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BounceToHand{Target: target}.Apply(ctx)
}
