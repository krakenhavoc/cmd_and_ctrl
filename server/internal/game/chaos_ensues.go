package game

import (
	"github.com/google/uuid"
)

// chaos_ensues.go is "chaos ensues" said by a resolving spell or ability
// (CR 311.7, #5648). The rule, quoted from the September 25, 2026 edition
// AGENTS.md §6 pins:
//
//	311.7. Each plane card has a triggered ability that triggers
//	       "Whenever chaos ensues." These are called chaos abilities.
//	       Each one is indicated by a chaos symbol to the left of the
//	       ability, though the symbol itself has no special rules
//	       meaning. This ability triggers if the chaos symbol is rolled
//	       on the planar die (see rule 901.9b), if a resolving spell or
//	       ability says that chaos ensues, or if a resolving spell or
//	       ability states that chaos ensues for a particular object. In
//	       the last case, the chaos ability can trigger even if that
//	       plane card is still in the planar deck but revealed. A chaos
//	       ability is controlled by the current planar controller.
//
// The planar die's chaos face stays what it was, an EventRollDie whose
// PlanarFace is PlanarFaceChaos. The second case is EventChaosEnsues: no
// die is rolled, so nothing that watches a roll sees one (Ichor Elixir
// does not replace it, "whenever you roll one or more dice" does not
// trigger). A chaos ability watches both (effects.WheneverChaosEnsues).
// The third case, chaos for one named object, has no card yet.
//
// In a game with no planar deck there is no plane and so no chaos
// ability: the instruction does nothing, and nothing is announced.

// EventChaosEnsues — a resolving spell or ability said that chaos ensues
// (CR 311.7). Actor is the controller of that spell or ability, Source
// the card it came from. Emitted only at a table playing Planechase.
const EventChaosEnsues EventKind = "chaos_ensues"

// ChaosEnsuesForEffect is "chaos ensues" as an effect's instruction (CR
// 311.7): every face-up plane's chaos ability triggers, controlled by
// the planar controller. `player` controls the effect and `source` is
// its card. It does nothing at a table that is not playing Planechase.
//
// Caller must hold g.mu (write).
func (g *Game) ChaosEnsuesForEffect(player, source uuid.UUID) {
	if g.Planechase == nil || !g.playerInGameLocked(player) {
		return
	}
	ev := Event{Kind: EventChaosEnsues, Actor: player, Source: source}
	if c, ok := g.LookupCardForEffect(source); ok {
		ev.PublicNames = map[uuid.UUID]string{source: c.Name}
	}
	g.EmitEvent(ev)
}
