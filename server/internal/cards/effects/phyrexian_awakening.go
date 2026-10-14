package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Awakening — Enchantment {2}{W} (March of the Machine):
//
//	"When this enchantment enters, incubate 4.
//	 Phyrexians you control have vigilance."
//
// The grant is a layer-6 tribal keyword grant (TribalKeywordGrant),
// read through the effective subtypes, so a transformed Incubator (a
// Phyrexian Token) and a changeling both get it. Vigilance only means
// something on a creature, so the filter's creature requirement drops
// nothing printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0566ab5d-c804-4ac0-add2-87fc352b4c7e",
		Name:         "Phyrexian Awakening",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Phyrexian Awakening — incubate 4", Do(Incubate{N: 4})),
		},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Phyrexian"}, YoursOnly: true}, "vigilance"),
		},
	})
}
