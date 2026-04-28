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
// S19 sub-PR 3 migrates the ETB half from S14's OnETB direct-call
// to the Triggered slot. The dies-trigger ships in sub-PR 4
// (`s19-dies-triggers`) — keeping the two halves separate keeps
// the per-PR diff focused on one event kind.
//
// Sandbox parity with S14:
//   - OptionalPrompt drives the "you may" gate (S14 treated may as
//     do; S19 lets the controller decline if the basics pile is
//     known empty or they want to skip).
//   - SearchLibrary picks the first basic in library order
//     (predicate IsBasicLand) and enters it tapped via
//     TappedOnEntry — matching the literal card text.
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
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				ctx := NewContext(g, nil)
				_ = SearchLibrary{
					Player:        source.Controller,
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
				}.Apply(ctx)
				return nil
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Solemn Simulacrum — search for a basic land?",
			},
		}},
	})
}
