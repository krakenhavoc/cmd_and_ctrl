package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kismet — "Artifacts, creatures, and lands your opponents control
// enter the battlefield tapped."
//
// First S17 enters-tapped replacement. Watches RepEventMove with
// NewZone = ZoneBattlefield; AppliesTo gates on:
//   - the entering card's controller ≠ Kismet's controller
//     (opponent's permanent — so Kismet doesn't tap your own cards)
//   - the entering card's type line includes one of Artifact,
//     Creature, or Land (Kismet leaves enchantments + planeswalkers
//     alone)
//
// The entering card is looked up via LookupCardForEffect. For
// spells resolving off the stack the card may still be in transit
// when AppliesTo fires — the lookup tolerates both stack and
// battlefield zones because the pipeline runs pre-push but after
// the card has an InstanceID + controller stamped.
//
// Replace simply sets ev.EntersTapped = true. The battlefield-
// entry site (spell-resolve / land-play / MoveCardByIDAsCommander)
// reads this flag after the push and stamps Card.Tapped before
// EventETB fires.
func init() {
	Register(Spec{
		OracleID:     "81fdd1c4-d43b-4f8b-8712-7c2bf45a3e0b",
		Name:         "Kismet",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventMove {
						return false
					}
					if ev.NewZone != game.ZoneBattlefield {
						return false
					}
					entering, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return false
					}
					if entering.Controller == src.Controller {
						return false
					}
					return entering.IsCreature() || entering.IsArtifact() || entering.IsLand()
				},
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.EntersTapped = true
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Kismet: enter tapped",
			},
		},
	})
}
