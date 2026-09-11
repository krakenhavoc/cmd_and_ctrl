package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solemn Simulacrum ("Sad Robot") — 2/2 Artifact Creature — Golem
// for {4}:
//
//	"When Solemn Simulacrum enters the battlefield, you may search
//	your library for a basic land card, put that card onto the
//	battlefield tapped, then shuffle.
//	When Solemn Simulacrum dies, you may draw a card."
//
// S19 sub-PR 3 migrated the ETB half from S14's OnETB direct-call
// to the Triggered slot. Sub-PR 4 adds the dies half ("you may
// draw a card") as a second Triggered entry watching EventLTB —
// both halves now auto-fire off the same harvester. Each "yes"
// puts a real trigger on the stack; the search / draw happens on
// resolution.
//
// Sandbox parity with S14:
//   - OptionalPrompt drives the "you may" gate (S14 treated may as
//     do; S19 lets the controller decline if the basics pile is
//     known empty or they want to skip).
//   - The land enters tapped via TappedOnEntry, matching the
//     literal card text. S22: the controller picks WHICH basic;
//     the "you may" half is already covered by OptionalPrompt, so
//     the search itself is not marked Optional as well.
//   - Empty / no-basic library → SearchLibrary no-ops silently.
//
// OnETB is dropped — the listener now owns ETB dispatch for this
// card.
func init() {
	Register(Spec{
		OracleID: "00c0543c-2a1f-4425-8283-4062d74a1637",
		Name:     "Solemn Simulacrum",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Solemn Simulacrum — search for a basic land",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{
							Player:        item.Controller,
							Predicate:     IsBasicLand,
							Dest:          game.ZoneBattlefield,
							Limit:         1,
							Reveal:        true,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        "Solemn Simulacrum — a basic land",
						}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Solemn Simulacrum — search for a basic land?",
			},
		}, {
			// "When Solemn Simulacrum dies, you may draw a card."
			// S19 sub-PR 4 wires the dies half onto the LTB harvester
			// (cardDied gates graveyard-only — a bounced/exiled Solemn
			// does not draw). Optional, so it queues a prompt; drawing
			// always has an effect, so no HasLegalTarget predicate.
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Solemn Simulacrum — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Solemn Simulacrum — draw a card?",
			},
		}},
	})
}
