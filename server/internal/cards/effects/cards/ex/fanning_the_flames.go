package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fanning the Flames — Sorcery {X}{R}{R} (Stronghold, #5627):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Fanning the Flames deals X damage to any target."
//
// Blaze's announced X (CR 107.3a), dealt through the shared
// cmdrB19DealXToTheTarget body, with Capsize's buyback (CR 702.27a). The
// buyback {3} is added to the total cost on top of {X}{R}{R} (CR 601.2f);
// X = 0 is a legal cast that deals no damage. A target gone in response
// means the spell doesn't resolve (CR 608.2b) and goes to the graveyard,
// buyback paid or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1d5f0a1c-88e0-4603-97af-8ed47766812e",
		Name:          "Fanning the Flames",
		XMatters:      true,
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.DamageXToTarget(0)),
		Targets:       effects.TargetAny(),
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{3}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CmdrB19DealXToTheTarget(ctx)
		},
	})
}
