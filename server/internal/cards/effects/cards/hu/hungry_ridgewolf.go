package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hungry Ridgewolf — Creature — Wolf {1}{R}, 2/2 (Innistrad: Crimson Vow,
// #5495):
//
//	"As long as you control another Wolf or Werewolf, this creature gets
//	 +1/+0 and has trample."
//
// Two live statics on the source alone, both gated on some OTHER Wolf or
// Werewolf its controller controls, read live through the layers (so a
// changeling counts and the bonus ends the moment the last one leaves): a
// layer 7c +1/+0 and a layer 6 trample grant (Barrow Naughty's shape). The
// two type words are read as words of this card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a7cd356-c07b-47dd-b089-8ecdd0e15581",
		Name:         "Hungry Ridgewolf",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control another Wolf or Werewolf, this creature gets +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.VowB07SelfWhileAnotherWolf,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
				},
			},
			effects.KeywordGrant(effects.VowB07SelfWhileAnotherWolf, "trample"),
		},
	})
}
