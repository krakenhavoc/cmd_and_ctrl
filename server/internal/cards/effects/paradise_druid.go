package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paradise Druid — Creature — Elf Druid {1}{G}, 2/1:
//
//	"This creature has hexproof as long as it's untapped.
//	 {T}: Add one mana of any color."
//
// Hexproof is a layer 6 grant to itself while it is untapped, so it is
// lost the moment the Druid taps for mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e119869d-1f66-402b-8251-d351965cf0d3",
		Name:         "Paradise Druid",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && !target.Tapped
			}, "hexproof"), "This creature has hexproof as long as it's untapped."),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
