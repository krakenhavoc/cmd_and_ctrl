package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maul Splicer — Creature — Phyrexian Human Artificer {6}{G}, 1/1:
//
//	"When this creature enters, create two 3/3 colorless Phyrexian Golem
//	 artifact creature tokens.
//	 Golem creatures you control have trample."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0d0c534-8bd2-4801-aadd-dcc9c92d6ec3",
		Name:         "Maul Splicer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Maul Splicer — create two 3/3 colorless Phyrexian Golem artifact creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 2})),
		},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Golem"}, YoursOnly: true}, "trample"),
		},
	})
}
