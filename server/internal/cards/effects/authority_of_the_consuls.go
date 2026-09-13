package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Authority of the Consuls — "Creatures your opponents control
// enter tapped. Whenever a creature an opponent controls enters,
// you gain 1 life."
//
// Two independent halves on one card:
//   - The enters-tapped half is a CR 614 replacement, the Kismet
//     shape narrowed from artifact/creature/land to creature only.
//   - The lifegain half is an ordinary ETB trigger watching every
//     opponent's creature, so it goes on the stack and can be
//     responded to.
//
// Both are controller-relative: an opponent's creature is any
// creature whose controller is not this card's controller.
func init() {
	Register(Spec{
		OracleID:     "55f3c721-e13a-406e-bc8e-d6cdc91ac477",
		Name:         "Authority of the Consuls",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || entering.Controller == src.Controller {
					return false
				}
				return entering.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Authority of the Consuls: enter tapped",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				entering, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || entering.Controller == source.Controller {
					return false
				}
				return entering.IsCreature()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Authority of the Consuls — gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
