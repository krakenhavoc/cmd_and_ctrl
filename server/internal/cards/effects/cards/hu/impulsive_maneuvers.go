package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Impulsive Maneuvers — Enchantment {2}{R}{R}:
//
//	"Whenever a creature attacks, flip a coin. If you win the flip, the
//	 next time that creature would deal combat damage this turn, it
//	 deals double that damage instead. If you lose the flip, the next
//	 time that creature would deal combat damage this turn, prevent that
//	 damage."
//
// Every attacking creature, anyone's, triggers it; you flip. Desperate
// Gambit's two halves (NextTimeFlip) on the attacker, combat damage only:
// the win is a "next time" multiplier (ADR 0108 §3), the loss a combat-only
// next-damage shield. Non-combat damage from the creature passes both by
// and spends neither. An attacker that has left the battlefield by the
// time the trigger resolves is gone (CR 400.7): the coin is flipped and
// nothing else happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39a9323d-dddc-42ac-929d-3f4fa7c87567",
		Name:         "Impulsive Maneuvers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil
			}, "Impulsive Maneuvers — flip a coin for the attacking creature", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				flip := effects.NextTimeFlip{CombatOnly: true, Question: "Impulsive Maneuvers — call the coin flip"}
				attacker := ctx.Trigger().Event.CardID
				if effects.OnBattlefield(g, attacker) {
					if info, ok := ctx.TriggeringPermanent(); !ok || !info.Left {
						if ref, zone, ok := g.DamageSourceRefLocked(attacker); ok {
							flip.Source, flip.SourceZone = ref, zone
						}
					}
				}
				return flip.Apply(ctx)
			}),
		},
	})
}
