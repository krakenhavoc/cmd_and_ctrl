package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gond Gate — Land — Gate (Commander Legends: Battle for Baldur's Gate, S82
// GAP #5948):
//
//	"Gates you control enter untapped.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color that a Gate you control could produce."
//
// The entry clause is Spelunking's replacement narrowed to Gates (CR 614.12:
// the entering permanent is read as it will exist, so its Gate type counts):
// it clears the entering Gate's "enters tapped", the Gate's own or another
// effect's, for its controller. The mana ability is Reflecting Pool's
// derivation over the Gates the controller controls, colours only ({C} is
// not a colour, CR 105.1): a Gate whose own ability is derived is followed
// (CR 106.7), the source never reads itself, and with no Gate offering a
// colour it adds nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4306938b-c0db-4e63-a4fb-61628e5ff41f",
		Name:         "Gond Gate",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || !ev.EntersTapped || src == nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == src.Controller && effects.CardHasSubtype(c, "Gate")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = false
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Gond Gate: Gates you control enter untapped",
		}},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: effects.ManaAbilityCost{Tap: true},
				DerivedMatch: func(c game.Card, controller uuid.UUID) bool {
					return c.Controller == controller && effects.CardHasSubtype(c, "Gate")
				},
				DerivedColorsOnly:       true,
				DerivesFromOtherSources: true,
				Label:                   "Add one mana of any color that a Gate you control could produce",
			},
		},
	})
}
