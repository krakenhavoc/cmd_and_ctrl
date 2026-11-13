package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Contamination — Enchantment {2}{B} (Urza's Saga):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless
//	 you sacrifice a creature.
//	 If a land is tapped for mana, it produces {B} instead of any other
//	 type and amount."
//
// The upkeep tax is The Gitrog Monster's over creatures: choose to
// sacrifice a creature or Contamination, and with no creature to offer
// Contamination goes.
//
// The land half is Damping Sphere's CR 614 replacement on the mana
// production event (CR 106.12b) without the "two or more" condition:
// every land TAPPED for mana, anyone's, produces exactly one {B}, so an
// Ancient Tomb's {C}{C} and a Forest's {G} are each {B}. Mana from a land
// that isn't tapped for it (a sacrificed land's ability with no {T}) is
// not affected. The player whose land is producing the mana orders this
// against other production replacements (CR 616.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4013b4c2-c9ed-4d14-90f0-97214ec1fded",
		Name:         "Contamination",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Contamination — sacrifice this unless you sacrifice a creature",
				sacrificeThisUnlessYouSacrificeA("Contamination", "a creature", Creature(), sacrificeACreatureOrThisAnswer)),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventManaAdded},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventProduceMana || !ev.ManaFromTap || len(ev.ManaColors) == 0 {
					return false
				}
				land, ok := g.LookupCardForEffect(ev.ManaSource)
				return ok && land.IsLand()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.ManaColors = []string{"B"}
				return nil
			},
			Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
				return ev.ManaPlayer
			},
			Label: "Contamination: a land tapped for mana produces {B}",
		}},
	})
}
