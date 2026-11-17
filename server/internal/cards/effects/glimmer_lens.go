package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bb7cebb7-a979-4879-b834-deab13f2a22c",
		Name:         "Glimmer Lens",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ForMirrodin("Glimmer Lens"),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureAttacked(ev, source) &&
					attackersYouControlOtherThan(g, ev.Actor, ev.CardID) >= 1
			}, "Glimmer Lens — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}{W}")},
	})
}
