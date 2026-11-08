package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindsparker — Creature — Elemental {1}{R}{R}, 3/2 (Foundations, #4465):
//
//	"First strike
//	 Whenever an opponent casts a white or blue instant or sorcery spell,
//	 this creature deals 2 damage to that player."
//
// Gleeful Arsonist's cast trigger with a fixed 2 and a narrower spell: an
// instant or sorcery that is white, blue or both, read off the spell on
// the stack. It triggers on the cast, so it resolves first and happens
// even if the spell is countered. "That player" is the caster, read off
// the item's carried event. A Mindsparker that has left the battlefield
// in response still deals the damage, named as the departed object
// (DealDamage.SourceObject). First strike rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "54866c25-b332-46c8-a2c5-1a71b702ddc0",
		Name:            "Mindsparker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnOpponentCast(And(Or(OfColor("W"), OfColor("U")), Or(Instant(), Sorcery()))),
				"Mindsparker — 2 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					caster := ctx.Trigger().Event.Actor
					ref, ok := ctx.SourceRef()
					if !ok || caster == uuid.Nil {
						return nil
					}
					return DealDamage{SourceObject: &ref, Target: caster, Amount: 2}.Apply(ctx)
				}),
		},
	})
}
