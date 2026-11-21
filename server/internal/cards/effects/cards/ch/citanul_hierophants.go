package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const citanulHierophantsGrant = "citanul-hierophants/tap-for-green"

// Citanul Hierophants — Creature — Human Druid {3}{G}, 3/2:
//
//	"Creatures you control have "{T}: Add {G}.""
//
// Jaheira's granted "{T}: Add {G}" (TapForManaGrant) on Cryptolith Rite's
// subject, every creature you control, the Hierophants included. A
// granted {T} ability is a creature's own, so summoning sickness applies
// to it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f67df84f-667e-41cf-80f9-911abafe70ac",
		Name:         "Citanul Hierophants",
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{effects.TapForManaGrant(citanulHierophantsGrant, "{G}", "Add {G}", "{T}: Add {G}.")},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(effects.B16CreaturesYouControl, citanulHierophantsGrant), "Creatures you control have \"{T}: Add {G}.\""),
		},
	})
}
