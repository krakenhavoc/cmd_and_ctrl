package game

import "github.com/google/uuid"

// echo.go — ADR 0108 §5 (#1888): the one fact echo reads.
//
//	702.30a Echo is a triggered ability. "Echo [cost]" means "At the
//	        beginning of your upkeep, if this permanent came under your
//	        control since the beginning of your last upkeep, sacrifice it
//	        unless you pay [cost]."
//
// "Came under your control since the beginning of your last upkeep" is
// two counters and a comparison:
//
//   - Player.UpkeepsBegun goes up by one as that player's upkeep begins,
//     before any "at the beginning of your upkeep" trigger is checked
//     (runStepEntryHooksLocked).
//   - Card.ControlledSinceUpkeep is set to the controller's UpkeepsBegun
//     whenever the permanent comes under that player's control: as it
//     enters (stampBattlefieldEntryLocked) and at each control change
//     the layer pass materialises (materialiseControlLocked).
//
// The trigger event is the beginning of your upkeep U, so "your last
// upkeep" is the one before it, U-1. A permanent stamped with U-1 or
// later came under your control after upkeep U-1 began. Hence the
// condition is ControlledSinceUpkeep >= UpkeepsBegun-1.
//
// Counting upkeeps rather than turns is what makes the edge cases fall
// out. A permanent that enters in your untap step is stamped U-1 and is
// charged at upkeep U, once. A skipped upkeep is never begun, so it is
// not a "last upkeep". An extra turn's upkeep is one more upkeep.
//
// Nothing here is stored beyond the two counters. The "Echo due" marker
// the table sees is derived from them (EchoDueLocked).

// KeywordEcho is the machine-readable name an echo trigger carries in
// TriggeredAbility.Keyword (effects.Echo stamps it).
const KeywordEcho = "echo"

// upkeepsBegunForLocked is Player.UpkeepsBegun for `player`, or 0 for a
// player who is not seated. Caller must hold g.mu.
func (g *Game) upkeepsBegunForLocked(player uuid.UUID) int {
	if p := g.playerByIDLocked(player); p != nil {
		return p.UpkeepsBegun
	}
	return 0
}

// UpkeepsBegunFor is upkeepsBegunForLocked for already-locked card
// callbacks.
func (g *Game) UpkeepsBegunFor(player uuid.UUID) int {
	return g.upkeepsBegunForLocked(player)
}

// CameUnderControlSinceLastUpkeepForEffect reports CR 702.30a's
// intervening "if" for the permanent `cardID` and the player `player`:
// the permanent is on the battlefield, `player` controls it, and it came
// under their control since the beginning of their last upkeep.
//
// A permanent `player` does not control answers false. Echo's
// consequence is "sacrifice it", and a player can't sacrifice a
// permanent they don't control (CR 701.21a), so a trigger whose
// permanent was taken away in response does nothing.
//
// Caller must hold g.mu.
func (g *Game) CameUnderControlSinceLastUpkeepForEffect(cardID, player uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	if c == nil || c.Controller != player {
		return false
	}
	return c.ControlledSinceUpkeep >= g.upkeepsBegunForLocked(player)-1
}

// AnnounceEchoPaidForEffect emits EventEchoPaid for the permanent
// `cardID` whose echo cost `payer` has just paid: the event Shah of Naar
// Isle's "When this creature's echo cost is paid" watches. Called from
// the echo prompt's "yes" branch, once the payment has been made.
//
// Caller must hold g.mu.
func (g *Game) AnnounceEchoPaidForEffect(cardID, payer uuid.UUID) {
	g.EmitEvent(Event{Kind: EventEchoPaid, Actor: payer, Source: cardID, CardID: cardID})
}

// EchoDueLocked reports whether the battlefield permanent `c` has echo
// and its echo will trigger at its controller's next upkeep: the
// table's "Echo due" marker (ADR 0108 §5 decision 5).
//
// At the controller's next upkeep their UpkeepsBegun will be one more
// than it is now, so the trigger's condition then is
// ControlledSinceUpkeep >= UpkeepsBegun (as it is now). A permanent
// whose echo was just charged in this upkeep was stamped before it and
// is no longer due.
//
// Caller must hold g.mu.
func (g *Game) EchoDueLocked(c *Card) bool {
	if c == nil || c.Controller == uuid.Nil {
		return false
	}
	if c.ControlledSinceUpkeep < g.upkeepsBegunForLocked(c.Controller) {
		return false
	}
	for _, t := range TriggersForCard(*c) {
		if t.Keyword == KeywordEcho {
			return true
		}
	}
	return false
}
