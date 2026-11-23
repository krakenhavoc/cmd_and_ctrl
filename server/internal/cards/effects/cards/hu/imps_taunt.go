package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imps' Taunt — Instant {1}{B} (Tempest, #5628):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Target creature attacks this turn if able."
//
// Into the Fray's attack requirement (AddAttackRequirementMod, checked
// as attackers are declared, CR 508.1d) until end of turn, with
// Capsize's buyback (CR 702.27a). It binds only the creature's
// controller, and only if that creature can attack this turn: cast on a
// creature whose controller isn't the active player, or after attackers
// are declared, it does nothing. Nothing is destroyed if it doesn't
// attack. A target gone in response means the spell doesn't resolve
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7ebec169-d941-460e-9e3c-ba54b5066b05",
		Name:          "Imps' Taunt",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{3}")},
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Imps' Taunt — attacks this turn if able",
			}.Apply(ctx)
		})),
	})
}
