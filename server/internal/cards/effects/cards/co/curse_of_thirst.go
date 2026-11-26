package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Thirst — Enchantment — Aura Curse {4}{B} (Dark Ascension, S82 OLD,
// #5571):
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, this Aura deals damage to
//	 that player equal to the number of Curses attached to them."
//
// Curse of Surveillance's upkeep trigger (the player whose upkeep it is, when
// that is the player this Curse is attached to), with the Curses counted as
// the ability resolves, this one included; the damage is the Aura's, read as
// it last was if it has left (CR 608.2h). With no Curse left attached, no
// damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffe3ad06-50d8-4ca3-a1f7-aa31dc2c486e",
		Name:         "Curse of Thirst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			},
			Key: "Curse of Thirst — this Aura deals damage to that player equal to the number of Curses attached to them",
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Trigger.Event.Actor
				n := effects.S82OldB440CursesAttachedToPlayer(g, player)
				if n <= 0 {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: n}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
