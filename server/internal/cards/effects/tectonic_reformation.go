package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tectonic Reformation — Enchantment {1}{R}:
//
//	"Each land card in your hand has cycling {R}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A hand grant (#3233, ADR 0093's amendment of 2026-10-13) of the
// ordinary Cycling row to land cards, beside the enchantment's own
// printed cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10da681b-50bd-4a1b-9934-f0c90e0af55c",
		Name:         "Tectonic Reformation",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Grants: []AbilityGrant{{
			Key:       "tectonic-reformation/cycling",
			Zone:      game.ZoneHand,
			Activated: []ActivatedAbility{Cycling("{R}")},
			Text:      "Cycling {R}",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneHand, CardsYouOwn(landCard), "tectonic-reformation/cycling"),
				"Each land card in your hand has cycling {R}."),
		},
	})
}

// landCard is the "land card" of a zone grant.
func landCard(card *game.Card) bool { return card.IsLand() }
