package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dakkon Blackblade — Legendary Creature — Human Warrior {2}{W}{U}{U}{B}, */*
// (Secret Lair Drop):
//
//	"Dakkon Blackblade's power and toughness are each equal to the number of
//	 lands you control."
//
// The characteristic-defining ability over lands (CR 604.3, layer 7a), the
// one Molimo and Beanstalk Giant share.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1aa78e3b-bf7e-4645-9521-4b3537c93f46",
		Name:         "Dakkon Blackblade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LandsYouControlPT("Dakkon Blackblade's power and toughness are each equal to the number of lands you control."),
		},
	})
}
