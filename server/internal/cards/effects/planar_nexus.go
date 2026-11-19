package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planar Nexus — Land (Modern Horizons 3 Commander, #5948):
//
//	"This land is every nonbasic land type. (Nonbasic land types include
//	 Cave, Desert, Gate, Lair, Locus, Mine, Power-Plant, Sphere, Tower, and
//	 Urza's.)
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Dryad of the Ilysian Grove's layer-4 append, narrowed to this land and
// widened to every land type that is not one of the five basic ones (CR
// 205.3i): Cave, Desert, Gate, Lair, Locus, Mine, Planet, Power-Plant,
// Sphere, Tower, Town and Urza's. A card with a Desert or Locus count sees
// it as one, and the types are the effective ones, so a Planar Nexus that
// loses its types by an effect stops counting. The second mana ability
// costs {1} and {T} as a mana ability (CR 605.1a), so it is a net gain of
// colour and no mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26005003-afcb-4c32-a760-be950246ff0f",
		Name:         "Planar Nexus",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This land is every nonbasic land type.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, want := range game.LandTypes {
					if game.IsBasicLandType(want) {
						continue
					}
					found := false
					for _, st := range c.Subtypes {
						if st == want {
							found = true
							break
						}
					}
					if !found {
						c.Subtypes = append(c.Subtypes, want)
					}
				}
			},
		}},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
	})
}
