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
//
// Reviewed against the oracle text for #1306. "A basic land card" is
// read off the Basic supertype (b30IsBasicLandCard) rather than the
// shared IsBasicLand's "basic land" type-line substring, which misses
// a Snow-Covered basic ("Basic Snow Land — Forest") — a real basic
// land card the printed search can find.
//
// One engine-wide posture applies here as on every optional trigger
// in the catalog: the "you may" is asked as the trigger is put on the
// stack rather than as it resolves (ADR 0018). Declining puts nothing
// on the stack, which is the same outcome as declining on resolution.
func init() {
	Register(Spec{
		OracleID:     "00c0543c-2a1f-4425-8283-4062d74a1637",
		Name:         "Solemn Simulacrum",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Solemn Simulacrum — search for a basic land", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     b30IsBasicLandCard,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Solemn Simulacrum — a basic land",
				}.Apply(NewContext(g, item))
			}), "Solemn Simulacrum — search for a basic land?"),
			// "When Solemn Simulacrum dies, you may draw a card."
			// S19 sub-PR 4 wires the dies half onto the LTB harvester
			// (cardDied gates graveyard-only — a bounced/exiled Solemn
			// does not draw). Optional, so it queues a prompt; drawing
			// always has an effect, so no HasLegalTarget predicate.
			Optional(WhenThisDies("Solemn Simulacrum — draw a card", Do(DrawCards{N: 1})), "Solemn Simulacrum — draw a card?"),
		},
	})
}
