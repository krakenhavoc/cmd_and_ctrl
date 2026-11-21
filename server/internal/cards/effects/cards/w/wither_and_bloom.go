package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "cf1e1d8d-c02e-4b66-b1db-49ec48fa4b03",
		Name:         "Wither and Bloom",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.BoostUntilEOT{Target: id, Power: -3, Toughness: -3, Label: "Wither and Bloom — -3/-3 until end of turn"}.Apply(ctx)
		})),
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}, Exile this card from your graveyard: Put a +1/+1 counter on target creature you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{B}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:       effects.RenewCountersOnTarget(game.CounterPlusOne, 1),
		}},
	})
}
