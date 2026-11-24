package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seraphic Greatsword — Artifact — Equipment {1}{W} (Commander Legends):
//
//	"Equipped creature gets +2/+2.
//	 Whenever equipped creature attacks the player with the most life or tied
//	 for most life, create a 4/4 white Angel creature token with flying
//	 that's tapped and attacking that player.
//	 Equip {4}"
//
// The pump is the usual attached static. The trigger is the equipped
// creature's attack declaration at a player who leads (or ties for) life
// (CR 508.3a: an attack at a planeswalker or battle is no attack on that
// player); the token is created tapped and attacking that same player, so
// nothing triggers for its attack (CR 508.4). A defender who has left the game
// by resolution gets no token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14a19102-00ed-4425-8837-1d05bc7ad8ba",
		Name:         "Seraphic Greatsword",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source) &&
					g.ClassifyAttackTargetForEffect(ev.Target) == game.AttackTargetPlayer &&
					effects.PlayerHasTheMostLife(g, ev.Target)
			},
			Key: "Seraphic Greatsword — create a 4/4 white Angel creature token with flying that's tapped and attacking that player",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Seraphic Greatsword — create a 4/4 white Angel creature token with flying that's tapped and attacking that player")
				item.Params.Player = ev.Target
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				defender := item.Params.Player
				if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
					return nil
				}
				tmpl := effects.TokenCard("4/4 white Angel with flying")
				tmpl.Tapped = true
				tmpl.AttackingTarget = defender
				return effects.CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
