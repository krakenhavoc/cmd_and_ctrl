package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Buried Treasure — Artifact — Treasure {2}:
//
//	"{T}, Sacrifice this artifact: Add one mana of any color.
//	 {5}, Exile this card from your graveyard: Discover 5. Activate only
//	 as a sorcery."
//
// The first ability is a Treasure's. The second works from the
// GRAVEYARD (CR 113.6, embalm's shape): the card exiles itself as the
// cost. Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "56f5ed8d-ef53-4766-9ef3-1e24a10e267b",
		Name:         "Buried Treasure",
		Completeness: CompletenessFull,
		Discovers:    true,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{5}, Exile this card from your graveyard: Discover 5",
			Cost:         Plus(ManaCost("{5}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       DiscoverN(5),
		}},
	})
}
