package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oblivion Sower — Creature — Eldrazi {6}, 5/8 (EDHREC rank 2315):
//
//	"When you cast this spell, target opponent exiles the top four
//	 cards of their library, then you may put any number of land
//	 cards that player owns from exile onto the battlefield under your
//	 control."
//
// The Eldrazi that steals lands out of exile. The cast trigger is a
// FromStack ability (cascade's mechanism, Kozilek's shape): it fires
// when the spell is announced, targets an opponent as it goes on the
// stack, and resolves above the Sower, so a countered Sower still
// exiles and still takes the lands. The four cards are an EXILE from
// the library, not a mill, so no mill payoff sees them; then every
// land card that player owns in exile — the four just exiled and any
// exiled earlier by anything else, as printed — is offered to the
// controller, who picks any number of them (#1872: "you may put any
// number" was taken at its maximum before, with no prompt). The picks
// enter under the controller's control as ONE entry, each as a new
// object with its own enters-tapped clause and ETB triggers, so each
// sees the others enter (CR 603.6a).
//
// A card exiled face down has no characteristics (CR 406.3a), so a
// foretold land is not a land card and is not offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d39b9f64-dc9b-413f-8d05-e21ef46d6756",
		Name:         "Oblivion Sower",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Oblivion Sower — target opponent exiles four, you take their lands from exile",
			// The cast trigger's controller is the CASTER (ev.Actor),
			// not source.Controller — a controller override the
			// engine cannot derive, so Build fills it in and leaves
			// item.Effect nil (ADR 0041 P9's fill-in Build).
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return &game.StackItem{
					Kind:         game.StackItemTriggered,
					Controller:   ev.Actor,
					Owner:        ev.Actor,
					SourceCardID: source.InstanceID,
					Label:        "Oblivion Sower — target opponent exiles four, you take their lands from exile",
				}
			},
			Effect: b21ExileTopFourThenTakeTheirLands,
		}},
	})
}
