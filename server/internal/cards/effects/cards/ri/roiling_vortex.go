package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roiling Vortex — Enchantment {1}{R}:
//
//	"At the beginning of each player's upkeep, this enchantment deals 1
//	 damage to them.
//	 Whenever a player casts a spell, if no mana was spent to cast that
//	 spell, this enchantment deals 5 damage to that player.
//	 {R}: Your opponents can't gain life this turn."
//
// The upkeep ping hits whoever's upkeep it is, the controller
// included. The free-spell punisher is Vexing Bauble's trigger with
// damage in place of a counter: the intervening if (CR 603.4) is read
// off the spell's own payment record (NoManaWasSpentToCast) when the
// ability triggers and again as it resolves, and "that player" is the
// caster the cast event names. The activation is ADR 0107 §5's "this
// turn" grant (#1880), covering the activator's opponents until
// cleanup.
//
// One caveat, Vexing Bauble's: a cast the engine did not charge
// (an insufficient permissive pool or Cast anyway) asks the caster what
// they paid outside the pool, or whether they cast it without paying its
// mana cost (#3855): the Vortex declares ReadsEveryCast. A known free
// cast in either payment mode is punished.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "34f83b7a-45a7-40a8-abcb-31d7aae4cea1",
		Name:           "Roiling Vortex",
		ReadsManaSpent: game.ReadsEveryCast,
		Completeness:   effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Roiling Vortex — 1 damage to that player", func(g *game.Game, item *game.StackItem) error {
				if p := effects.TriggeringActor(item); p != uuid.Nil {
					return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: 1}.Apply(effects.NewContext(g, item))
				}
				return nil
			}),
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && effects.NoManaWasSpentToCast(g, ev.CardID)
			}, "Roiling Vortex — 5 damage to a player who cast a spell for no mana", func(g *game.Game, item *game.StackItem) error {
				p := effects.TriggeringActor(item)
				// CR 603.4: the intervening if, again on resolution.
				if p == uuid.Nil || !effects.NoManaWasSpentToCast(g, item.Trigger.Event.CardID) {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: 5}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Your opponents can't gain life this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PlayersCantGainLifeThisTurn{Opponents: true, Label: "Roiling Vortex — your opponents can't gain life this turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
