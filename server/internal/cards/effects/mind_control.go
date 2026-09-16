package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Control — Enchantment — Aura for {3}{U}{U}:
//
//	"Enchant creature
//	 You control enchanted creature."
//
// Deferred three times before this: ADR 0012 sent it to S17, ADR 0013
// sent it on to S24 on attachment grounds, and ADR 0036 left it out
// of the attachment spike because layer 2 was a stub and
// Card.Controller was written directly in a dozen places. The layer
// listener carried the note in its own source
// (`layer_listener.go`: "Mind Control / aura attach is deferred").
//
// The card itself is two lines, because all of the work was in the
// engine: an enchant clause and a layer-2 continuous effect. What
// makes the two-line version correct rather than a fake is that
// control REVERTS on its own. Nothing here remembers the previous
// controller and nothing has to hand it back — the effect simply
// stops being in the active set when the Aura leaves, and the next
// recompute re-seeds from the creature's control baseline.
//
// Everything that follows from a control change follows for free
// too, because the recompute materialises layer 2's answer onto
// Card.Controller before anything reads it:
//
//   - The creature attacks and blocks for its new controller, and
//     can be tapped for their costs.
//   - CR 302.6: it has summoning sickness under the new controller
//     until their untap step, so stealing a creature does not hand
//     over an attack this turn (unless it has haste).
//   - CR 506.4: stealing a creature mid-combat removes it from
//     combat.
//   - It moves to the new controller's panel on the board, because
//     the client partitions on CardView.controller.
//   - CR 704.5m still applies: if the creature stops being a
//     creature, the Aura goes to the graveyard and control reverts
//     in the same state-based-action pass.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5912546a-acc2-448c-b042-64bdac5ec129",
		Name:         "Mind Control",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			ControlAttachedBySource(),
		},
	})
}
