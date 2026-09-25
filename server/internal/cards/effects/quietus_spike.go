package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quietus Spike — Artifact — Equipment for {3} (EDHREC rank 2750):
//
//	"Equipped creature has deathtouch.
//	 Whenever equipped creature deals combat damage to a player, that
//	 player loses half their life, rounded up.
//	 Equip {3}"
//
// Four connections kill a player from forty, and the deathtouch is
// what makes the connections happen: a 1/1 with the Spike trades with
// anything that blocks it, so the defender's good block is a bad
// block.
//
// The halving is life LOSS, not damage. That distinction is printed
// and it matters: it is not prevented, not redirected, and not
// reduced, and it does not care about the creature's power. It reads
// the damaged player's life at RESOLUTION — captured nowhere — so a
// lifegain trigger that resolves above this one makes the halving
// bigger.
//
// "Rounded up" is the ceiling division `(life + 1) / 2`, which is
// why 40 becomes 20 but 21 becomes 11. On a player at 1, it takes 1.
// A player already at or below zero loses nothing, because half of a
// non-positive number rounded up is not a loss and the engine's
// state-based action has them anyway.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ff0a309-9b4a-4f9e-9ea6-c4d07906ebec",
		Name:         "Quietus Spike",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("deathtouch"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Quietus Spike — halve that player's life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Quietus Spike — halve that player's life")
				item.Params.Player = ev.Target
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				damaged := item.Params.Player
				p := g.PlayerByIDForEffect(damaged)
				if p == nil || p.Life <= 0 {
					return nil
				}
				half := (p.Life + 1) / 2
				return GainLife{Player: damaged, Amount: -half}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
