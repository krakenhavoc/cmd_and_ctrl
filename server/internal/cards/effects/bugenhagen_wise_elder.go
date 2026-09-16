package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bugenhagen, Wise Elder — Legendary Creature — Human Shaman {1}{G},
// 1/3 (EDHREC rank 2560):
//
//	"Reach
//	 At the beginning of your upkeep, if you control a creature with
//	 power 7 or greater, draw a card.
//	 {T}: Add one mana of any color."
//
// A two-mana rainbow dork that pays off a big creature. Reach rides
// PrintedKeywords; the mana ability is Birds of Paradise's five-way
// pick at its printed width; the upkeep trigger's intervening-if —
// a creature with current power 7 or greater under the controller's
// control, counters and anthems included — is checked as the
// trigger would fire and again as it resolves (CR 603.4), so a
// giant that dies in response draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4228dd75-1fdd-48dc-9259-b841b2e46c64",
		Name:            "Bugenhagen, Wise Elder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 7)
			}, "Bugenhagen, Wise Elder — draw a card", func(g *game.Game, item *game.StackItem) error {
				if !b24ControlsCreatureWithPowerAtLeast(g, item.Controller, 7) {
					return nil
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
