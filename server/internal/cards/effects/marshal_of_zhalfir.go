package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marshal of Zhalfir — Creature — Human Knight {W}{U}, 2/2 (March of the
// Machine):
//
//	"Other Knights you control get +1/+1.
//	 {W}{U}, {T}: Tap another target creature."
func init() {
	Register(Spec{
		OracleID:     "9ad01fb2-2ba7-447a-a5ef-e57d46421635",
		Name:         "Marshal of Zhalfir",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}, {T}: Tap another target creature.",
			Cost:    Plus(ManaCost("{W}{U}"), TapCost()),
			Targets: Another(TargetCreature("another target creature")),
			Effect:  tapChosenPermanent,
		}},
	})
}
