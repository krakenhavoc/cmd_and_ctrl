package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mnemonic Sliver — Creature — Sliver {2}{U}, 2/2 (Tempest, S82 OLD batch 766,
// #5628):
//
//	"All Slivers have '{2}, Sacrifice this permanent: Draw a card.'"
//
// Basal Sliver's shape: an ability bundle granted to every Sliver on the
// battlefield, whoever controls it (CR 613.1f). Each Sliver's own sacrifice is
// its own cost, "this permanent" being the Sliver that has the ability.
//
// No simplification.
func init() {
	const grantKey = "mnemonic-sliver/draw"
	effects.Register(effects.Spec{
		OracleID:     "101a5c2d-b0b4-45a7-b236-df474ac47356",
		Name:         "Mnemonic Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}, Sacrifice this permanent: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect:  effects.B36DrawOne,
			}},
			Text: "{2}, Sacrifice this permanent: Draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, grantKey),
		},
	})
}
