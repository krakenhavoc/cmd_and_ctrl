package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Maro — Creature — Elemental {4}{G}, */*:
//
//	"Domain — Territorial Maro's power and toughness are each equal to
//	 twice the number of basic land types among lands you control."
//
// A characteristic-defining ability (CR 604.3, layer 7a, CR 613.4a) on
// the shared domain count (DomainPT, #5666).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "478eeed5-8170-4fb3-8ec2-bcebc0265622",
		Name:         "Territorial Maro",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.DomainPT("Domain — Territorial Maro's power and toughness are each equal to twice the number of basic land types among lands you control", 2, 2),
		},
	})
}
