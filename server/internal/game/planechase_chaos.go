package game

import "github.com/google/uuid"

// planechase_chaos.go is the "chaos ensues" instruction (ADR 0216 slice 4,
// #4870, Oteclán): the second way a chaos ability triggers (CR 311.7).
//
//	"Each plane card has a triggered ability that triggers “Whenever
//	 chaos ensues.” ... This ability triggers if the chaos symbol is
//	 rolled on the planar die (see rule 901.9b), if a resolving spell or
//	 ability says that chaos ensues, or if a resolving spell or ability
//	 states that chaos ensues for a particular object."
//
// EventChaosEnsues is the event a resolving "chaos ensues" instruction
// emits. It is NOT a planar die roll, so nothing that watches a roll of the
// planar die (No Way Out's "whenever you roll the planar die") sees it.

// EventChaosEnsues — a resolving spell or ability said that chaos ensues
// (CR 311.7). Actor is the player whose spell or ability it was. Additive
// within the event schema.
const EventChaosEnsues EventKind = "chaos_ensues"

// ChaosEnsuesForEffect is "chaos ensues": every face-up plane's chaos
// ability triggers (they are controlled by the planar controller, CR 901.6).
// A table without Planechase has no plane to trigger, so it does nothing.
//
// Caller must hold g.mu (write).
func (g *Game) ChaosEnsuesForEffect(actor uuid.UUID) {
	if g.Planechase == nil {
		return
	}
	g.EmitEvent(Event{Kind: EventChaosEnsues, Actor: actor})
}
