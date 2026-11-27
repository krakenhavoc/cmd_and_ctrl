package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overlaid Terrain — Enchantment {2}{G}{G} (Nemesis, S82 OLD batch 679, #5618):
//
//	"As this enchantment enters, sacrifice all lands you control.
//	 Lands you control have "{T}: Add two mana of any one color.""
//
// The sacrifice happens as the enchantment enters (CR 614.12), not as a
// trigger, so nothing can respond to it and the Terrain's own grant is not yet
// in play: the lands are sacrificed together, as one event, as the lands they
// were. Then every land its controller controls, including one that enters
// later, has the layer-6 grant (ADR 0093) of Mul Daya Channelers' mana ability:
// two mana of one colour, chosen on the tap.
//
// No simplification.
func init() {
	const grantKey = "overlaid-terrain/two-of-one-color"
	const grantText = "{T}: Add two mana of any one color."
	effects.Register(effects.Spec{
		OracleID:     "e5818ab3-8c37-4063-9e9d-cbe4ae0bba14",
		Name:         "Overlaid Terrain",
		Completeness: effects.CompletenessFull,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			owner := card.Controller
			return effects.B08SacrificeAllMatching(ctx, effects.And(effects.Land(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.Controller == owner
			}))
		},
		Grants: []effects.AbilityGrant{{
			Key:  grantKey,
			Text: grantText,
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: effects.OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{effects.Labelled(effects.GrantAbilities(effects.LandsYouControl, grantKey), "Lands you control have \"{T}: Add two mana of any one color.\"")},
	})
}
