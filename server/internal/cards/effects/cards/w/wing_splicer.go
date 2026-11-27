package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wing Splicer — Creature — Phyrexian Human Artificer {3}{U}, 1/1:
//
//	"When this creature enters, create a 3/3 colorless Phyrexian Golem
//	 artifact creature token.
//	 Golem creatures you control have flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fbf6afc-8a16-447b-bdef-b7d0b640f76c",
		Name:         "Wing Splicer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wing Splicer — create a 3/3 colorless Phyrexian Golem artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1})),
		},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Golem"}, YoursOnly: true}, "flying"),
		},
	})
}
