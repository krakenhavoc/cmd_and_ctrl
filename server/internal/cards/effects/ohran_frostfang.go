package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ohran Frostfang — Snow Creature — Snake {3}{G}{G}, 2/6 (EDHREC rank
// ~296):
//
//	"Attacking creatures you control have deathtouch.
//	 Whenever a creature you control deals combat damage to a player,
//	 draw a card."
//
// A card-advantage anthem: every attacker becomes a trade the
// defender cannot decline, and any of them connecting refills the
// hand. "Snow" is cosmetic here — nothing on this card reads it.
//
// # The static is the seam this card proves (#1218)
//
// "Attacking creatures" is a Layer 6 keyword grant whose AppliesTo
// reads Card.AttackingTarget — the one input to a continuous effect
// that nothing invalidated on before this fix. Hand size (#74) and
// life total (#1117) were the other two non-battlefield inputs a
// static can read; attacking status was the "Still open" quarter
// #1117 left (docs/engine-seams.md, "Layer cache invalidation on
// hand, life, attack and graveyard events"). Without the bump, an
// attacker declared after the last recompute would show as vanilla
// until some unrelated event happened to invalidate the cache, and a
// creature pulled out of combat mid-declaration (a control change) or
// left attacking when combat ended would keep the keyword it no
// longer earns.
//
// DependsOnAttackingStatus declares the dependency; the fix that
// reads it is layer_listener.go's EventAttack arm and
// invalidateLayersForAttackChangeLocked (called from
// removeFromCombatLocked and clearCombatLocked, the two exits with no
// event of their own). See internal/game/layer_invalidation_attack_test.go
// for the engine-level proof and TestOhranFrostfangGrantsDeathtouchOnlyToAttackers
// below for this card's.
//
// The drain half is ordinary machinery: combatDamageToPlayerBy is the
// same predicate Bident of Thassa and Coastal Piracy use, mandatory
// here rather than a "you may".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b99ada26-9a61-4175-9fb8-15a106960220",
		Name:         "Ohran Frostfang",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:                    game.Layer6Ability,
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.AttackingTarget != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !keywordSliceContains(c.Abilities, "deathtouch") {
					c.Abilities = append(c.Abilities, "deathtouch")
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Ohran Frostfang — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
