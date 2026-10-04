package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gleeful Arsonist — Creature — Human Wizard {2}{R}, 1/2:
//
//	"Whenever an opponent casts a noncreature spell, this creature
//	 deals damage equal to its power to that player.
//	 Undying (When this creature dies, if it had no +1/+1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 +1/+1 counter on it.)"
//
// The cast trigger is Kambal, Consul of Allocation's: on the CAST, so it
// resolves above the spell and happens even if the spell is countered.
// "That player" is the caster, read off the item's carried event.
//
// "Its power" is read as the ability RESOLVES (CR 608.2h): the
// Arsonist's power with its counters, so the +1/+1 counter undying puts
// on it makes it hit for 2. An Arsonist that has left the battlefield in
// response still deals the damage, equal to its last-known power, and
// the damage names the departed object (DealDamage.SourceObject), so a
// returned Arsonist is a new object whose power is not the one read
// (CR 400.7).
//
// Undying is PrintedKeywords; the engine derives the trigger
// (game/undying_persist.go, #2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9025b3ce-f6ae-431d-9eb5-02d713b8c9b8",
		Name:            "Gleeful Arsonist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: AnOpponentCast(Noncreature()),
			Key:       "Gleeful Arsonist — damage equal to its power to that player",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				caster := ctx.Trigger().Event.Actor
				ref, ok := ctx.SourceRef()
				if !ok || caster == uuid.Nil {
					return nil
				}
				arsonist, ok := ctx.SourcePermanent()
				if !ok || arsonist.Power <= 0 {
					return nil
				}
				return DealDamage{SourceObject: &ref, Target: caster, Amount: arsonist.Power}.Apply(ctx)
			},
		}},
	})
}
