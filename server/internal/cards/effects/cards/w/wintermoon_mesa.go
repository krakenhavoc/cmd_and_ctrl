package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wintermoon Mesa — Land (Prophecy, S82 OLD batch 675, #5617):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Tap two target lands."
//
// The enters-tapped replacement (CR 614.1d) and a mana ability (CR 605.1a) as
// every tapland has them. The third ability is an ordinary activated ability
// (it targets, so it is not a mana ability): {2}, {T} and the land itself are
// the cost, paid before it resolves, and it needs exactly two target lands,
// any player's (Nomad Decoy's two-target tap). Each target is rechecked as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4a6f95e-856c-4eb5-82ba-b2406be22b23",
		Name:         "Wintermoon Mesa",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Tap two target lands.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
			Effect:  effects.Msc13TapTheLegalTargets,
		}},
	})
}
