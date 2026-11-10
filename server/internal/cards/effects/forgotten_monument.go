package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forgotten Monument — Land — Cave (The Lost Caverns of Ixalan, #4470):
//
//	"{T}: Add {C}.
//	 Other Caves you control have '{T}, Pay 1 life: Add one mana of any
//	 color.'"
//
// The grant is the ordinary ADR 0093 layer-6 bundle (Cryptolith Rite's
// shape) with Mana Confluence's row: the life is a COST, so a Cave can't
// be tapped for it at 0 life (CR 119.4). "Other Caves you control" is
// read live off each permanent after the type-changing layers, so a land
// that becomes a Cave gets the row and one that stops being a Cave loses
// it. The Monument itself is not one of them. "Cave" is a word of the
// Monument's text, read through it (CR 612.2).
//
// No simplification.
const forgottenMonumentGrant = "forgotten-monument/pay-life-any-color"

func init() {
	Register(Spec{
		OracleID:      "71393988-ad6f-43fd-9978-c0de15ae8e87",
		Name:          "Forgotten Monument",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{colorlessTapRow()},
		Grants: []AbilityGrant{{
			Key: forgottenMonumentGrant,
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Life: 1},
				Produced: "{W|U|B|R|G}",
				Label:    "Pay 1 life: Add one mana of any color",
			}},
			Text: "{T}, Pay 1 life: Add one mana of any color.",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.HasSubtype(source.TextSubtype("Cave"))
			}, forgottenMonumentGrant), "Other Caves you control have \"{T}, Pay 1 life: Add one mana of any color.\""),
		},
	})
}
