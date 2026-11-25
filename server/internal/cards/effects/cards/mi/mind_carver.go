package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Carver — Artifact — Equipment {B} (Zendikar Rising, #5508):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+0. It gets +3/+1 instead as long as an
//	 opponent has eight or more cards in their graveyard.
//	 Equip {2}{B}"
//
// One layer 7c static whose amount is read live: +3/+1 while any opponent of
// the Equipment's controller has eight cards in their graveyard, +1/+0 when
// none does (Expedition Lookout's read), never both. The enters trigger
// attaches it as it resolves, to the target chosen as it went on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65d6eae4-4318-4dbe-af7a-9f57bb587605",
		Name:         "Mind Carver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Equipped creature gets +1/+0. It gets +3/+1 instead as long as an opponent has eight or more cards in their graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if effects.S82OpponentHasGraveyardCards(g, source.Controller, 8) {
					c.Power += 3
					c.Toughness++
					return
				}
				c.Power++
			},
		}},
		Triggered: []game.TriggeredAbility{effects.AfrB06EnterAttach("Mind Carver")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{B}")},
	})
}
