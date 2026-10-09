package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scab-Clan Berserker — Creature — Human Berserker {1}{R}{R}, 2/2:
//
//	"Haste
//	 Renown 1 (When this creature deals combat damage to a player, if
//	 it isn't renowned, put a +1/+1 counter on it and it becomes
//	 renowned.)
//	 Whenever an opponent casts a noncreature spell, if this creature
//	 is renowned, this creature deals 2 damage to that player."
//
// #2049: renown is the engine's keyword trigger (game/renown.go). The
// cast trigger is Gleeful Arsonist's: on the CAST, so it resolves above
// the spell, and "that player" is the caster. "If this creature is
// renowned" is an intervening if (CR 603.4), read as the spell is cast
// and again as the trigger resolves, from the Berserker's last-known
// information if it has left; the damage then names the departed
// object as its source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4f659ba-c277-47a2-8b95-ce81ddac34fb",
		Name:            "Scab-Clan Berserker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "renown 1"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, sourceLKI game.Characteristic, g *game.Game) bool {
				return ThisIsRenowned(source) && AnOpponentCast(Noncreature())(ev, source, sourceLKI, g)
			},
			Key: "Scab-Clan Berserker — 2 damage to that player",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				caster := ctx.Trigger().Event.Actor
				ref, ok := ctx.SourceRef()
				if !ok || caster == uuid.Nil || !ThisWasRenowned(ctx) {
					return nil
				}
				return DealDamage{SourceObject: &ref, Target: caster, Amount: 2}.Apply(ctx)
			},
		}},
	})
}
