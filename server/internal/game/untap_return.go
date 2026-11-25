package game

import (
	"fmt"

	"github.com/google/uuid"
)

// untap_return.go — "During your next untap step, as you untap your
// permanents, return this land to its owner's hand" (Undiscovered
// Paradise, #6218; CR 502.3).
//
// The untap step has no triggers (CR 502.4: no player receives
// priority), so this is not a delayed trigger: it is an action of the
// untap step itself, taken as the player's permanents untap. A record,
// one ScopedEffect (ScopeGame, IndefiniteDuration) with one
// ModReturnAtNextUntap mod naming the player whose untap step it waits
// for (Player) and the permanent (Objects[0], pinned to the object it
// was, CR 400.7). performUntapStepLocked spends every record of the
// active player as it determines the untap set: a permanent still that
// object goes to its owner's hand, one that has left is gone and the
// record does nothing. A skipped untap step never reaches it, so the
// record waits for the next untap step that does happen (CR 614.10a).
// A record whose object has left is dropped then too, whoever it waits
// for. An older binary refuses a file carrying one (the unknown kind).

// ModReturnAtNextUntap is "during <Player>'s next untap step, as they
// untap their permanents, return <Objects[0]> to its owner's hand"
// (#6218). Scope ScopeGame, IndefiniteDuration. Reads Player and
// Objects (exactly one).
const ModReturnAtNextUntap ModKind = "returnAtNextUntap"

// returnAtNextUntapModProblem is the registration and restore check.
// "" when the mod is sound or is another kind.
func returnAtNextUntapModProblem(m Mod) string {
	if m.Kind != ModReturnAtNextUntap {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a returnAtNextUntap mod names no player"
	}
	if len(m.Objects) != 1 || m.Objects[0].ID == uuid.Nil {
		return "a returnAtNextUntap mod must name exactly one object"
	}
	return ""
}

// ReturnAtNextUntapForEffect records that the permanent `cardID` returns
// to its owner's hand during `player`'s next untap step, as they untap
// their permanents (Undiscovered Paradise). Returns false, writing
// nothing, when the card is not on the battlefield or the player is not
// seated or has left.
//
// Safe inside a resolution or a mana ability: it only writes a record.
//
// Caller must hold g.mu (write).
func (g *Game) ReturnAtNextUntapForEffect(cardID, player uuid.UUID, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return false
	}
	m := Mod{Kind: ModReturnAtNextUntap, Player: player, Objects: []ObjectRef{{ID: cardID, Epoch: c.ObjectEpoch}}}
	if problem := returnAtNextUntapModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	return g.RegisterScopedRuleEffectForEffect(cardID, ScopeGame, uuid.Nil, []Mod{m}, IndefiniteDuration(), label)
}

func returnAtNextUntapRecord(e ScopedEffect) (Mod, bool) {
	if len(e.Mods) != 1 || e.Mods[0].Kind != ModReturnAtNextUntap {
		return Mod{}, false
	}
	return e.Mods[0], true
}

// liveRefLocked reports the battlefield permanent `ref` names, nil when
// that object has left.
func (g *Game) liveRefLocked(ref ObjectRef) *Card {
	if c := findBattlefieldCard(g, ref.ID); c != nil && c.ObjectEpoch == ref.Epoch {
		return c
	}
	return nil
}

// spendReturnsAtUntapLocked runs in `activePlayer`'s untap step, before
// the untap set is read: every record waiting on this step is used up
// and its permanent, if it is still that object, returns to its owner's
// hand. Records whose object has left are dropped too.
//
// Caller must hold g.mu (write).
func (g *Game) spendReturnsAtUntapLocked(activePlayer uuid.UUID) {
	var due []ObjectRef
	for _, e := range g.ScopedEffects {
		if m, ok := returnAtNextUntapRecord(e); ok && m.Player == activePlayer {
			due = append(due, m.Objects[0])
		}
	}
	g.dropScopedEffectsLocked(func(_ int, e ScopedEffect) bool {
		m, ok := returnAtNextUntapRecord(e)
		return ok && (m.Player == activePlayer || g.liveRefLocked(m.Objects[0]) == nil)
	})
	for _, ref := range due {
		if g.liveRefLocked(ref) == nil {
			continue
		}
		_ = g.BounceToHandForEffect(ref.ID)
	}
}

// ReturnsAtNextUntapForLocked is how many records wait to return `cardID`
// at an untap step, for the view and the tests.
//
// Caller must hold g.mu (read or write).
func (g *Game) ReturnsAtNextUntapForLocked(cardID uuid.UUID) int {
	n := 0
	for _, e := range g.ScopedEffects {
		if m, ok := returnAtNextUntapRecord(e); ok && m.Objects[0].ID == cardID && g.liveRefLocked(m.Objects[0]) != nil {
			n++
		}
	}
	return n
}
