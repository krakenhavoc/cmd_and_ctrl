package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimmer Lens — Artifact — Equipment {1}{W} (Phyrexia: All Will Be One,
// Commander):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Whenever equipped creature and at least one other creature attack,
//	 draw a card.
//	 Equip {1}{W}"
//
// #5479, S82. The battalion shape (Arthur, Marigold Knight): every
// attacker is staged before the declaration is announced, so the equipped
// creature's own EventAttack already sees the others. EventAttack fires
// once per attacker, so the equipped creature is one trigger per combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb7cebb7-a979-4879-b834-deab13f2a22c",
		Name:         "Glimmer Lens",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ForMirrodin("Glimmer Lens"),
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source) &&
					effects.AttackersYouControlOtherThan(g, ev.Actor, ev.CardID) >= 1
			}, "Glimmer Lens — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}{W}")},
	})
}
