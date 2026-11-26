package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, the Firebrand — Legendary Planeswalker — Chandra {3}{R}, loyalty 3
// (Magic 2013, S82 OLD, #5568):
//
//	"+1: Chandra deals 1 damage to any target.
//	 −2: When you next cast an instant or sorcery spell this turn, copy that
//	 spell. You may choose new targets for the copy.
//	 −6: Chandra deals 6 damage to each of up to six targets."
//
// The +1 is the shared target damage. The −2 is Complete the Circuit's delayed
// trigger for the next instant or sorcery cast this turn, copying it once with
// the new-targets offer (CR 707.10c): a spell cast before the ability resolves is
// not copied, and the trigger is spent by the first one. The −6 deals its damage
// as one event to each target still legal (CR 120.2), up to six, players and
// planeswalkers included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "536c1112-694f-42f8-a7b1-604389e54a95",
		Name:            "Chandra, the Firebrand",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Chandra deals 1 damage to any target.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.DealDamageToTheTarget(1),
			},
			{
				Label: "−2: When you next cast an instant or sorcery spell this turn, copy that spell. You may choose new targets for the copy.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.WhenYouNextCast(
						"Chandra, the Firebrand — copy that spell",
						game.CastFilter{Types: []string{"Instant", "Sorcery"}},
						effects.CopyTheSpellBody,
					).Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−6: Chandra deals 6 damage to each of up to six targets.",
				Cost:    effects.LoyaltyCost(-6),
				Targets: effects.TargetAny().WithCount(0, 6),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 6)),
				Effect:  effects.S82OldB430SixDamageToEachTarget,
			},
		},
	})
}
