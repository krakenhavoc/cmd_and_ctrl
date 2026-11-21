package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2965bae9-0d2b-42e1-b735-01ea57a72409",
		Name:         "Tangled Skyline",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tangled Skyline — you gain 5 life and incubate 5",
				effects.Do(effects.GainLife{Amount: 5}, effects.Incubate{N: 5})),
		},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Phyrexian"}, YoursOnly: true}, "reach"),
		},
	})
}
