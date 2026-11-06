package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zurgo's Vanguard — Creature — Dog Soldier {2}{R}, */3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the
//	 beginning of the next end step.)
//	 Zurgo's Vanguard's power is equal to the number of creatures you
//	 control."
//
// Mobilize is mobilize.go (CR 702.181). The power is Ironroot Warlord's
// characteristic-defining ability (layer 7a), the Vanguard itself
// counted, read live: the mobilized Warrior raises it by one while it is
// on the battlefield, before combat damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6fdfd94-cef5-482c-83ed-11c5ba40abe9",
		Name:         "Zurgo's Vanguard",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			jmpB5PowerIsCreaturesYouControl("Zurgo's Vanguard's power is equal to the number of creatures you control."),
		},
		Triggered: []game.TriggeredAbility{Mobilize("Zurgo's Vanguard", 1)},
	})
}
