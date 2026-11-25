package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mesmerizing Benthid — Creature — Octopus {3}{U}{U}, 4/5 (Ravnica
// Allegiance):
//
//	"When this creature enters, create two 0/2 blue Illusion creature tokens
//	 with "Whenever this token blocks a creature, that creature doesn't untap
//	 during its controller's next untap step."
//	 This creature has hexproof as long as you control an Illusion."
//
// The Illusion is a catalog token that carries its own freeze trigger
// (OldB181IllusionToken); the hexproof is a static while any Illusion is on
// your side, the Benthid itself being an Octopus.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f36d3d87-5596-4c26-b8d0-f2979a166b57",
		Name:         "Mesmerizing Benthid",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.HasSubtype("Illusion")), "hexproof"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mesmerizing Benthid — create two 0/2 blue Illusion creature tokens",
				effects.Do(effects.CreateToken{Template: effects.OldB181IllusionToken(), N: 2})),
		},
	})
}
