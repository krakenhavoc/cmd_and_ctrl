package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:      "7ebec169-d941-460e-9e3c-ba54b5066b05",
		Name:          "Imps' Taunt",
		Completeness:  CompletenessFull,
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Imps' Taunt — attacks this turn if able",
			}.Apply(ctx)
		})),
	})
}
