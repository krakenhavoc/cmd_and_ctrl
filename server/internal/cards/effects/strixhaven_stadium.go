package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strixhaven Stadium — Artifact {3}:
//
//	"{T}: Add {C}. Put a point counter on this artifact.
//	 Whenever a creature deals combat damage to you, remove a point counter
//	 from this artifact.
//	 Whenever a creature you control deals combat damage to an opponent,
//	 put a point counter on this artifact. Then if it has ten or more point
//	 counters on it, remove them all and that player loses the game."
//
// The mana ability's counter is a Rider, so it is part of the mana ability
// (CR 605.1a). Each creature that connects triggers on its own (the
// printed rulings), so neither damage trigger batches. "That player" is
// the opponent the damage event names; the ten-counter check happens as
// the ability resolves, after the counter is put (a tenth counter added
// any other way loses no one).
func init() {
	const point = "point"
	Register(Spec{
		OracleID:     "2ed0c6fc-b8b4-47da-a04b-1d995bdeecb0",
		Name:         "Strixhaven Stadium",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}. Put a point counter on this artifact.",
			Rider: func(g *game.Game, _ uuid.UUID, source uuid.UUID) error {
				return g.AddCounterForEffect(source, point, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, strixhavenCreatureDealtCombatDamageToYou,
				"Strixhaven Stadium — remove a point counter from this artifact",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if !onBattlefield(g, ctx.Source()) || counterCountOn(g, ctx.Source(), point) == 0 {
						return nil
					}
					return AddCounter{Target: ctx.Source(), Kind: point, N: -1}.Apply(ctx)
				}),
			On(game.EventDealDamage, strixhavenYourCreatureDealtCombatDamageToAnOpponent,
				"Strixhaven Stadium — put a point counter on this artifact; then if it has ten or more, remove them all and that player loses the game",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if !onBattlefield(g, ctx.Source()) {
						return nil
					}
					if err := (AddCounter{Target: ctx.Source(), Kind: point, N: 1}).Apply(ctx); err != nil {
						return err
					}
					n := counterCountOn(g, ctx.Source(), point)
					if n < 10 || item.Trigger == nil {
						return nil
					}
					if err := (AddCounter{Target: ctx.Source(), Kind: point, N: -n}).Apply(ctx); err != nil {
						return err
					}
					return LoseTheGame{Player: item.Trigger.Event.Target}.Apply(ctx)
				}),
		},
	})
}

// strixhavenCreatureDealtCombatDamageToYou — a creature (any controller's)
// dealt combat damage to the source's controller.
func strixhavenCreatureDealtCombatDamageToYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Target != source.Controller {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.IsCreature()
}

// strixhavenYourCreatureDealtCombatDamageToAnOpponent — a creature the
// source's controller controls dealt combat damage to an opponent.
func strixhavenYourCreatureDealtCombatDamageToAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Target != source.Controller && combatDamageToPlayerBy(ev, source.Controller, g)
}
