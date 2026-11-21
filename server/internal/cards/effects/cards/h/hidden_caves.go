package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Lost Caverns of Ixalan's five mono-coloured Caves, one table:
//
//	"This land enters tapped.
//	 {T}: Add {C}.   (the land's colour)
//	 {4}{C}, {T}, Sacrifice this land: Discover 4. Activate only as a
//	 sorcery."
//
// Hidden Cataract (U), Hidden Courtyard (W), Hidden Necropolis (B),
// Hidden Nursery (G) and Hidden Volcano (R). The enters-tapped clause is
// the self-replacement every tapland uses, so nothing watching for a tap
// sees one. The discover is ADR 0099's (game/discover.go).
func init() {
	for _, t := range []struct {
		oracleID, name, color string
	}{
		{"927979d7-9b5c-4448-aef0-baf2907a89f1", "Hidden Cataract", "U"},
		{"e19d5071-4ea1-4883-b067-a21e553f96e0", "Hidden Courtyard", "W"},
		{"f780ee53-62b0-4c32-b5b7-047651f48e5f", "Hidden Necropolis", "B"},
		{"1a26e2d6-6bfc-4cdc-9bd6-8b37a9be2961", "Hidden Nursery", "G"},
		{"a1c7cd7a-0795-4135-b787-effeb981d95b", "Hidden Volcano", "R"},
	} {
		mana := "{" + t.color + "}"
		effects.Register(effects.Spec{
			OracleID:     t.oracleID,
			Name:         t.name,
			Completeness: effects.CompletenessFull,
			Discovers:    true,
			Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
			ManaAbilities: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: mana,
				Label:    "Add " + mana,
			}},
			Activated: []effects.ActivatedAbility{{
				Label:        "{4}" + mana + ", {T}, Sacrifice this land: Discover 4",
				Purpose:      game.Purpose{Discover: 4},
				Cost:         effects.Plus(effects.ManaCost("{4}"+mana), effects.TapCost(), effects.SacrificeThis()),
				SorcerySpeed: true,
				Effect:       effects.DiscoverN(4),
			}},
		})
	}
}
