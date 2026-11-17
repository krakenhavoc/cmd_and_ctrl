package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Norn's Decree — Enchantment {2}{W}:
//
//	"Whenever one or more creatures an opponent controls deal combat
//	 damage to you, that opponent gets a poison counter.
//	 Whenever a player attacks, if one or more players being attacked are
//	 poisoned, the attacking player draws a card."
//
// The first trigger is Nelly Borca's per-dealer batch (CR 603.2c): one
// trigger for each opponent whose creatures dealt you combat damage in a
// damage step, however many creatures. The dealing creature's controller
// is the damage event's Actor, stamped when the damage is dealt. The
// counter is placed by the trigger's controller through the CR 614
// player-counter window.
//
// The second is one trigger per attack declaration (CR 603.2c, the
// Mangara batch). Its intervening "if" (CR 603.4) is read when the
// attack is declared and again as the trigger resolves: some creature the
// attacking player controls is still attacking a player (not a
// planeswalker or battle) with one or more poison counters. "The
// attacking player" draws, whoever controls Norn's Decree.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92c43ee1-0f7d-4993-8d3a-689cc87d7480",
		Name:         "Norn's Decree",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventDealDamage && ev.Target == source.Controller &&
						creatureDealtCombatDamageToAnOpponentOf(ev, uuid.Nil, g) &&
						ev.Actor != uuid.Nil && ev.Actor != source.Controller
				},
				Key:          "Norn's Decree — that opponent gets a poison counter",
				OncePerBatch: true,
				BatchKey: func(ev game.Event, _ *game.Card, _ *game.Game) string {
					return ev.Actor.String()
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					who := item.Trigger.Event.Actor
					if p := g.PlayerByIDForEffect(who); p == nil || p.Eliminated {
						return nil
					}
					return g.AddPlayerCounterByForEffect(item.Controller, who, game.CounterPoison, 1)
				},
			},
			OncePerBatch(On(game.EventAttack,
				func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && nornsDecreeAttackingAPoisonedPlayer(g, ev.Actor)
				},
				"Norn's Decree — the attacking player draws a card",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					who := item.Trigger.Event.Actor
					if !nornsDecreeAttackingAPoisonedPlayer(g, who) {
						return nil
					}
					return DrawCards{Player: who, N: 1}.Apply(NewContext(g, item))
				})),
		},
	})
}

// nornsDecreeAttackingAPoisonedPlayer is the intervening "if one or more
// players being attacked are poisoned": a creature `attacker` controls is
// attacking a player who has a poison counter.
func nornsDecreeAttackingAPoisonedPlayer(g *game.Game, attacker uuid.UUID) bool {
	for _, id := range b13AttackingCreaturesYouControl(g, attacker) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if p := g.PlayerByIDForEffect(c.AttackingTarget); p != nil && !p.Eliminated && p.Poison >= 1 {
			return true
		}
	}
	return false
}
