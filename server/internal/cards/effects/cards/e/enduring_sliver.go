package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enduring Sliver — Creature — Sliver {1}{W}, 2/2 (Modern Horizons):
//
//	"Outlast {2} ({2}, {T}: Put a +1/+1 counter on this creature. Outlast
//	 only as a sorcery.)
//	 Other Sliver creatures you control have outlast {2}."
//
// Outlast is the shared sorcery-speed tap-and-pay activation (CR 702.107a).
// The grant is a layer 6 ability grant over the other Slivers its controller
// controls, re-read on every recompute, read through the layered view so a
// changeling counts as a Sliver; the Sliver itself has outlast as printed.
// Opposing Slivers get nothing.
//
// No simplification.
func init() {
	const grant = "enduring-sliver/outlast"
	slivers := effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "f9726123-f87b-41bb-9b6f-6d2fbcf0eae2",
		Name:         "Enduring Sliver",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RprB04Outlast("{2}")},
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Text:      "Outlast {2}",
			Activated: []effects.ActivatedAbility{effects.RprB04Outlast("{2}")},
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.IsCreature() && slivers.Matches(target, g, source)
			}, grant),
		},
	})
}
