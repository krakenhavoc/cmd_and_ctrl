package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reclusive Taxidermist — {1}{G} Creature — Human Druid 1/2 (Innistrad:
// Crimson Vow, S82):
//
//	"This creature gets +3/+2 as long as there are four or more creature
//	 cards in your graveyard.
//	 {T}: Add one mana of any color."
//
// A layer 7c static that reads its controller's graveyard as it is now
// (Carrion Grub's reading; the engine recomputes whenever a card crosses
// a graveyard boundary) and the ordinary any-colour mana ability. A token
// never sits in a graveyard long enough to count (CR 111.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a0168c5-e258-4c50-96aa-95deb2198e6f",
		Name:         "Reclusive Taxidermist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			Label:     "This creature gets +3/+2 as long as there are four or more creature cards in your graveyard.",
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if n, _ := cmdr2S30CreatureCardsIn(g, source.Controller); n >= 4 {
					c.Power += 3
					c.Toughness += 2
				}
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
