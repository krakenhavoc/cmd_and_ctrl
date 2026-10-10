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
// The finality counter is two self-replacements on this card, because
// no other card in the catalog uses the counter yet: it enters with one
// only when it comes from a graveyard (a hard-cast Repriser has none),
// and a Repriser that carries one and would die is exiled instead
// (CR 614.6: the replacement rewrites the move, so no dies trigger
// sees it). A finality counter put on it some other way is honoured by
// the same second replacement.
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
			{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return ev.Kind == game.RepEventMove && ev.OldZone == game.ZoneBattlefield &&
						ev.NewZone == game.ZoneGraveyard && src != nil && ev.CardID == src.InstanceID &&
						src.Counters[rfCreatureBFinalityCounter] > 0
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.NewZone = game.ZoneExile
					ev.NewZoneOwner = uuid.Nil
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Grim Repriser: exile instead of dying (finality counter)",
			},
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
