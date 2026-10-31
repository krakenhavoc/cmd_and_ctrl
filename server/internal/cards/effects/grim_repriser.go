package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grim Repriser — Creature — Zombie Bard {B}{R}, 2/2:
//
//	"Prowess
//	 {B}{R}: Return this card from your graveyard to the battlefield
//	 with a finality counter on it. Activate only if an opponent has
//	 been dealt noncombat damage this turn. (If a creature with a
//	 finality counter on it would die, exile it instead.)"
//
// Reassembling Skeleton's graveyard activation, gated by an activation
// condition that reads this turn's events for noncombat damage dealt to
// one of the controller's opponents (rfCreatureBAnOpponentWasDealtNoncombatDamage).
// The finality counter is half the card's and half the engine's: the
// Repriser enters with one only when it comes from a graveyard (a
// hard-cast Repriser has none; a self-replacement), and a permanent
// that carries one and would die is exiled instead by the engine's own
// CR 122.1h replacement (game/finality_counters.go, #4575), which
// rewrites the move so no dies trigger sees it (CR 614.6). A finality
// counter put on it some other way is honoured the same way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e853e2fb-f90d-4934-aa70-c8d3cdf12e56",
		Name:            "Grim Repriser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Replacements: []game.ReplacementEffect{
			rfCreatureBEntersWithCounterFromGraveyard(rfCreatureBFinalityCounter, 1,
				"Grim Repriser: enters with a finality counter"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}{R}: Return this card from your graveyard to the battlefield with a finality counter on it. Activate only if an opponent has been dealt noncombat damage this turn.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{B}{R}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return rfCreatureBAnOpponentWasDealtNoncombatDamage(g, controller)
			},
			Effect: returnThisFromGraveyardToBattlefield,
		}},
	})
}
