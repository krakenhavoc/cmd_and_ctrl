package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Michiko Konda, Truth Seeker — Legendary Creature — Human Advisor
// {3}{W}, 2/2 (the Squall deck request, #3918):
//
//	"Whenever a source an opponent controls deals damage to you, that
//	 player sacrifices a permanent of their choice."
//
// One trigger per damage event (CR 603.2): three opponent's creatures
// connecting in one combat damage step are three events and three
// sacrifices, as the rulings say. Damage to your permanents does not
// count, only damage to you.
//
// "That player" is the controller of the source as the damage was dealt,
// read when the trigger fires (Elesh Norn's b17DamageSourceController)
// and carried on the item, so a source that changes hands or leaves
// before the trigger resolves still names the player who dealt it. They
// choose the permanent as it resolves; a player with no permanent
// sacrifices nothing.
//
// No simplification.
func init() {
	const label = "Michiko Konda — that player sacrifices a permanent"
	Register(Spec{
		OracleID:     "fbe190eb-fd4b-40e7-9a17-48c96b60f653",
		Name:         "Michiko Konda, Truth Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target != source.Controller {
					return false
				}
				dealer := b17DamageSourceController(ev, g)
				return dealer != uuid.Nil && dealer != source.Controller
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = b17DamageSourceController(ev, g)
				return item
			},
			Effect: michikoKondaSacrifice,
		}},
	})
}

// michikoKondaSacrifice has the damaging player sacrifice a permanent of
// their choice.
func michikoKondaSacrifice(g *game.Game, item *game.StackItem) error {
	player := item.Params.Player
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return SacrificeChoice{
		Player:     player,
		Candidates: PermanentsControlledBy(g, player),
		Question:   "Michiko Konda, Truth Seeker — sacrifice a permanent",
	}.Apply(NewContext(g, item))
}
