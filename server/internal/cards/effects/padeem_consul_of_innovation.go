package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Padeem, Consul of Innovation — Legendary Creature — Vedalken
// Artificer {3}{U}, 1/4 (slice 296-m):
//
//	"Artifacts you control have hexproof. (They can't be the targets
//	 of spells or abilities your opponents control.)
//	 At the beginning of your upkeep, if you control the artifact with
//	 the greatest mana value or tied for the greatest mana value, draw
//	 a card."
//
// The hexproof grant is a plain Layer 6 keyword grant over "artifacts
// you control" (b16GrantKeywords), the same shape every static
// keyword grant in the catalog uses.
//
// The upkeep draw is a CR 603.4 intervening-if: checked when the
// upkeep begins AND again on resolution (Colossal Majesty's pattern),
// so an artifact that stops being the biggest — sacrificed, or
// outsized by something that resolves in response — between trigger
// and resolution leaves no card. "The greatest mana value" reads
// every artifact on the battlefield, any controller's, and asks
// whether the upkeep player controls one that ties the max.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c7ba712-6a99-4d2f-9242-a2163a11f69c",
		Name:         "Padeem, Consul of Innovation",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsArtifact() && target.Controller == source.Controller
			}, "hexproof"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && padeemControlsGreatestArtifact(g, source.Controller)
			}, "Padeem, Consul of Innovation — draw a card", padeemDrawIfGreatestArtifact),
		},
	})
}

// padeemGreatestArtifactManaValue is the highest mana value among
// every artifact on the battlefield, any controller's. Zero when
// there are no artifacts in play, which no controller can tie.
func padeemGreatestArtifactManaValue(g *game.Game) (int, bool) {
	if g.Battlefield == nil {
		return 0, false
	}
	best := 0
	found := false
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsArtifact() {
			continue
		}
		if mv := c.ManaValue(); !found || mv > best {
			best = mv
			found = true
		}
	}
	return best, found
}

// padeemControlsGreatestArtifact reports whether `controller` controls
// an artifact tied for the greatest mana value among all artifacts on
// the battlefield.
func padeemControlsGreatestArtifact(g *game.Game, controller uuid.UUID) bool {
	best, found := padeemGreatestArtifactManaValue(g)
	if !found {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.IsArtifact() && c.Controller == controller && c.ManaValue() == best {
			return true
		}
	}
	return false
}

func padeemDrawIfGreatestArtifact(g *game.Game, item *game.StackItem) error {
	if !padeemControlsGreatestArtifact(g, item.Controller) {
		return nil
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}
