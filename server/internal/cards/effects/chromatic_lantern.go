package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chromatic Lantern — Artifact {3}:
//
//	"Lands you control have "{T}: Add one mana of any color."
//	 {T}: Add one mana of any color."
//
// A layer-6 grant to every land you control (ADR 0093). Because the
// grant is layer 6 and a basic-land-type change removes a land's own
// abilities in layer 4 (CR 305.7), a Lantern still fixes every land
// under Blood Moon — the rule's last sentence. A land's granted mana
// is an ordinary source for the auto-tapper (only a CREATURE's
// granted mana is kept for last).
//
// No simplification.
const chromaticLanternGrant = "chromatic-lantern/any-color"

func init() {
	Register(Spec{
		OracleID:     "539f5396-d99a-417d-a84c-dff7930b5900",
		Name:         "Chromatic Lantern",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Grants: []AbilityGrant{AnyColorManaGrant(chromaticLanternGrant)},
		Static: []game.StaticAbility{GrantAbilities(landsYouControl, chromaticLanternGrant)},
	})
}

// landsYouControl is the "lands you control" AppliesTo, read off the
// effective types so a land a layer-4 effect made is one.
func landsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsLand() && target.Controller == source.Controller
}
