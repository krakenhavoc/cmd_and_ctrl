package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unstable Amulet — Artifact {1}{R}:
//
//	"When this artifact enters, you get {E}{E} (two energy counters).
//	 Whenever you cast a spell from anywhere other than your hand, this
//	 artifact deals 1 damage to each opponent.
//	 {T}, Pay {E}{E}: Exile the top card of your library. You may play
//	 it until you exile another card with this artifact."
//
// The energy is ADR 0129's: a counter on the controller, paid as an
// activation cost component. The cast trigger is Vega, the Watcher's
// (b25CastFromNotHand): the cast event carries the zone the spell left,
// so a commander, a flashback, an impulse-exiled card (including one
// this Amulet exiled) all count. A copy is not cast and fires nothing.
//
// The exile is #2539's window (exileTopUntilYouExileAnother). The card
// stays playable until the activator exiles another card with this same
// Amulet object, and for as long as it stays exiled if the Amulet leaves
// the battlefield first (ruling 2024-06-07). Normal timing applies to
// what is played, and a land takes the turn's land play (ruling
// 2024-06-07).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3d4f5ab-0ff4-48e1-8cf0-163b290804a5",
		Name:         "Unstable Amulet",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Unstable Amulet", 2),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b25CastFromNotHand(ev, source.Controller)
			}, "Unstable Amulet — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:  "{T}, Pay {E}{E}: Exile the top card of your library. You may play it until you exile another card with this artifact.",
			Cost:   Plus(TapCost(), PayEnergy(2)),
			Effect: exileTopUntilYouExileAnother,
		}},
	})
}
