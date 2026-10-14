package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sculpted Perfection — Enchantment {2}{W}{B} (March of the Machine):
//
//	"When this enchantment enters, incubate 2.
//	 Phyrexians you control get +1/+1."
//
// A layer-7c tribal anthem (TribalAnthem) over the effective subtypes,
// so a transformed Incubator (a Phyrexian Token) is a 1/1 larger. A
// noncreature Phyrexian has no power or toughness to change, so the
// filter's creature requirement drops nothing printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90fa1303-ad1c-46da-a4d6-4b29ea49b7ca",
		Name:         "Sculpted Perfection",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sculpted Perfection — incubate 2", Do(Incubate{N: 2})),
		},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Phyrexian"}, YoursOnly: true}, 1, 1),
		},
	})
}
