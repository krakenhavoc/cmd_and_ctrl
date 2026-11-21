package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wither and Bloom — Instant {1}{B} (Modern Horizons 3, #5948):
//
//	"Target creature gets -3/-3 until end of turn.
//	 {1}{B}, Exile this card from your graveyard: Put a +1/+1 counter on
//	 target creature you control. Activate only as a sorcery."
//
// The spell is a shrink; the graveyard ability is Renew's shape
// (renew.go) without the ability word, so it is an ordinary activated
// ability that works from the graveyard and exiles the card as a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf1e1d8d-c02e-4b66-b1db-49ec48fa4b03",
		Name:         "Wither and Bloom",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return BoostUntilEOT{Target: id, Power: -3, Toughness: -3, Label: "Wither and Bloom — -3/-3 until end of turn"}.Apply(ctx)
		})),
		Activated: []ActivatedAbility{{
			Label:        "{1}{B}, Exile this card from your graveyard: Put a +1/+1 counter on target creature you control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{B}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature you control", YouControl()),
			Effect:       renewCountersOnTarget(game.CounterPlusOne, 1),
		}},
	})
}
