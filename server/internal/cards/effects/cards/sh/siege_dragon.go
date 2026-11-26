package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siege Dragon — Creature — Dragon {5}{R}{R}, 5/5 (Magic 2015, #5556):
//
//	"Flying
//	 When this creature enters, destroy all Walls your opponents control.
//	 Whenever this creature attacks, if defending player controls no Walls, it
//	 deals 2 damage to each creature without flying that player controls."
//
// The enters trigger is a mass destruction over the Walls opponents control
// (regeneration and indestructible apply as for any destroy). The attack
// trigger is Spectral Bears' shape: the defending player (CR 506.2) is fixed as
// it triggers and carried on the item, and the intervening "if" (CR 603.4) is
// checked as it triggers and again as it resolves. The damage is one instance
// to the creatures without flying that player controls then.
//
// No simplification.
func init() {
	const label = "Siege Dragon — 2 damage to each creature without flying defending player controls"
	attack := effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.AttackDeclared(ev, source) &&
			effects.S82OldB359DefenderControlsNoWall(g, g.DefendingPlayerForAttackForEffect(ev.Target))
	}, label, func(g *game.Game, item *game.StackItem) error {
		defender := item.Params.Player
		if defender == uuid.Nil || !effects.S82OldB359DefenderControlsNoWall(g, defender) {
			return nil
		}
		ctx := effects.NewContext(g, item)
		return g.DamageInstanceForEffect(func() error {
			return effects.DamageEachMatching(ctx,
				effects.And(effects.Creature(), effects.ControlledBy(defender), effects.WithoutKeyword("flying")), 2)
		})
	})
	attack.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = g.DefendingPlayerForAttackForEffect(ev.Target)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "206539a6-0e60-4986-bc57-c2b89e08b9f2",
		Name:            "Siege Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		// ADR 0126 §6: its attack trigger hits only the creatures without flying
		// of the player it attacks.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, OpponentsOnly: true, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Siege Dragon — destroy all Walls your opponents control",
				effects.Do(effects.DestroyAllMatching{Match: effects.And(effects.HasSubtype("Wall"), effects.OpponentControls())})),
			attack,
		},
	})
}
