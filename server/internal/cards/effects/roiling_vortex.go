package effects

import (
	"github.com/google/uuid"

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
// (strict mana off) records an UNKNOWN payment, and unknown is never
// "no mana", so the punisher only sees free casts at a strict-mana
// table. Weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "34f83b7a-45a7-40a8-abcb-31d7aae4cea1",
		Name:         "Roiling Vortex",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't track what was spent, so a spell cast for no mana isn't punished — turn strict mana on for that ability to work."},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Roiling Vortex — 1 damage to that player", func(g *game.Game, item *game.StackItem) error {
				if p := triggeringActor(item); p != uuid.Nil {
					return DealDamage{Source: item.SourceCardID, Target: p, Amount: 1}.Apply(NewContext(g, item))
				}
				return nil
			}),
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && NoManaWasSpentToCast(g, ev.CardID)
			}, "Roiling Vortex — 5 damage to a player who cast a spell for no mana", func(g *game.Game, item *game.StackItem) error {
				p := triggeringActor(item)
				// CR 603.4: the intervening if, again on resolution.
				if p == uuid.Nil || !NoManaWasSpentToCast(g, item.Trigger.Event.CardID) {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: p, Amount: 5}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label: "{R}: Your opponents can't gain life this turn.",
			Cost:  ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PlayersCantGainLifeThisTurn{Opponents: true, Label: "Roiling Vortex — your opponents can't gain life this turn"}.Apply(NewContext(g, item))
			},
		}},
	})
}
