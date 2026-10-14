package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangled Skyline — Enchantment {4}{G} (March of the Machine):
//
//	"When this enchantment enters, you gain 5 life and incubate 5.
//	 Phyrexians you control have reach."
//
// One enters trigger doing both, in printed order; the grant is a
// layer-6 tribal keyword grant (see Phyrexian Awakening).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2965bae9-0d2b-42e1-b735-01ea57a72409",
		Name:         "Tangled Skyline",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tangled Skyline — you gain 5 life and incubate 5",
				Do(GainLife{Amount: 5}, Incubate{N: 5})),
		},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Phyrexian"}, YoursOnly: true}, "reach"),
		},
	})
}
