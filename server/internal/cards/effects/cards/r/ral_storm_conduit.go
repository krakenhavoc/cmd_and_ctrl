package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ral, Storm Conduit — Legendary Planeswalker — Ral {2}{U}{R}, loyalty 4:
//
//	"Whenever you cast or copy an instant or sorcery spell, Ral deals 1 damage
//	 to target opponent or planeswalker.
//	 +2: Scry 1.
//	 −2: When you next cast an instant or sorcery spell this turn, copy that
//	 spell. You may choose new targets for the copy."
//
// The static trigger is magecraft's shape (CR 707.10: a copy is not cast, but
// this one watches both). −2 is a delayed trigger that fires once, on the next
// instant or sorcery you cast this turn, and copies from last-known
// information whether or not that spell is still on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7da26c9-78a6-41c5-a5ef-6555e8274d7a",
		Name:            "Ral, Storm Conduit",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Magecraft("Ral, Storm Conduit — 1 damage to target opponent or planeswalker", effects.DealDamageToTheTarget(1)),
				effects.TargetOpponentOrPlaneswalker()),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Scry 1.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Scry: 1},
				Effect:  effects.Do(effects.Scry{N: 1}),
			},
			{
				Label: "−2: When you next cast an instant or sorcery spell this turn, copy that spell. You may choose new targets for the copy.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: effects.Do(effects.WhenYouNextCast("Ral, Storm Conduit — copy that spell",
					game.CastFilter{Types: []string{"Instant", "Sorcery"}}, effects.CopyTheSpellBody)),
			},
		},
	})
}
