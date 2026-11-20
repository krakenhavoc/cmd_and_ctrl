package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Experimental Synthesizer — Artifact {R} (Kamigawa: Neon Dynasty, #5493):
//
//	"When this artifact enters or leaves the battlefield, exile the top card
//	 of your library. Until end of turn, you may play that card.
//	 {2}{R}, Sacrifice this artifact: Create a 2/2 white Samurai creature
//	 token with vigilance. Activate only as a sorcery."
//
// One ability with two trigger conditions (WhenThisEntersOrLeaves). The
// exile is an impulse draw whose window is the turn (ExileTopWithPermission's
// default): the card is playable, not just castable, so a land can be played.
// The second ability is Abzan Monument's sacrifice-to-token shape at sorcery
// speed; its own sacrifice is the leave-the-battlefield half.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11db299a-0918-46a3-b39b-16acc3f4e281",
		Name:         "Experimental Synthesizer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrLeaves("Experimental Synthesizer — exile the top card of your library; until end of turn, you may play that card",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.ExileTopWithPermission{From: ctx.Controller(), GrantTo: ctx.Controller(), N: 1}.Apply(ctx)
				}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{R}, Sacrifice this artifact: Create a 2/2 white Samurai creature token with vigilance. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.NeoB08SamuraiToken(), N: 1}),
		}},
	})
}
