package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mayhem Devil — 3/3 Creature — Devil for {1}{B}{R}:
//
//	"Whenever a player sacrifices a permanent, this creature deals 1
//	damage to any target."
//
// The first consumer of EventSacrifice (S21 sub-PR 1). Two things
// this card proves out:
//
//   - It triggers on ANY player's sacrifice and on ANY permanent,
//     not just creatures — cracking your own Treasure for mana pings
//     something, and so does an opponent's fetch-land-style sacrifice
//     if one ever lands in the catalog.
//   - Sacrifice is distinct from death. A creature destroyed by
//     Doom Blade does NOT trigger this, while a creature sacrificed
//     to Goblin Bombardment triggers both this and Blood Artist.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4709f11c-aef8-45ac-b2bf-e640c568dfac",
		Name:         "Mayhem Devil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventSacrifice},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != uuid.Nil
			},
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Key:     "Mayhem Devil — deal 1 damage",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.DealDamage{
					Source: ctx.Source(),
					Target: item.Targets[0].ID,
					Amount: 1,
				}.Apply(ctx)
			},
		}},
	})
}
