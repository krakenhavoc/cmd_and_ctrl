package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seven of Nine, Unassimilated — Legendary Artifact Creature — Human
// Officer {1}{U}, 0/4 (Star Trek, #4973):
//
//	"At the beginning of your first main phase, draw a card. Then discard
//	 a card unless you control five or more differently named lands.
//	 {6}{U}: Seven of Nine's base power and toughness become 7/9. Activate
//	 only if you control seven or more differently named lands."
//
// The trigger fires in the precombat main phase only (CR 505.1a). The
// lands are counted as it resolves: five or more differently named and it
// is a plain draw, fewer and it is a loot.
//
// The activated ability sets the BASE power and toughness in layer 7b
// (CR 613.4b), so counters and pumps still apply on top, and it states no
// duration, so it lasts as long as this is the same object (CR 611.2a,
// CR 400.7). The land count gates activation (CR 602.1b's "activate only
// if"), not resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af9c9b8a-0713-4904-8e86-2b3331ec5495",
		Name:         "Seven of Nine, Unassimilated",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Seven of Nine, Unassimilated — draw a card, then discard a card unless you control five or more differently named lands",
				func(g *game.Game, item *game.StackItem) error {
					if trekB8DifferentlyNamedLands(g, item.Controller) >= 5 {
						return DrawCards{N: 1}.Apply(NewContext(g, item))
					}
					return lootOne(g, item, 1)
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{6}{U}: Seven of Nine's base power and toughness become 7/9. Activate only if you control seven or more differently named lands.",
			Cost:  ManaCost("{6}{U}"),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return trekB8DifferentlyNamedLands(g, controller) >= 7
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     game.SetBasePTMods(7, 9),
					Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
					Label:    "Seven of Nine, Unassimilated — base power and toughness 7/9",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
