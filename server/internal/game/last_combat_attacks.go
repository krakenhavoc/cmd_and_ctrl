package game

import "github.com/google/uuid"

// last_combat_attacks.go — what a player's creatures did during that
// player's LAST COMBAT (Territorial Hellkite, #5368).
//
// last_turn_attacks.go answers "during your last TURN"; a combat phase is
// smaller than a turn (an extra combat is another one, a turn with no
// combat has none), so the creature that asks "an opponent this creature
// didn't attack during your last combat" needs its own record.
//
// Player.CombatAttacks collects the attacks declared in the combat phase
// in progress. As each combat phase of that player's turn begins (the
// step the beginning-of-combat triggers are harvested from, so it has
// rotated by the time they resolve), it becomes Player.LastCombatAttacks
// and the collection starts afresh. So during a beginning of combat step,
// LastCombatAttacks is the previous combat phase of that player — in the
// same turn when an extra combat was added, otherwise in an earlier turn
// — and an empty record is a real answer: that combat had no attackers.
// A turn whose combat is skipped begins no combat step and changes
// nothing. Only declarations are recorded (CR 508.4), keyed by object
// (CR 400.7).

// rotateCombatAttacksLocked starts the active player's new combat phase.
// Caller must hold g.mu.
func (g *Game) rotateCombatAttacksLocked() {
	seat := g.Turn.ActiveSeat
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	p := g.Seats[seat]
	// A fresh slice, never a shared backing array: the undo clone holds
	// the old record (the reason TurnTally.Attacks gives).
	p.LastCombatAttacks = append([]AttackRecord(nil), p.CombatAttacks...)
	p.CombatAttacks = nil
}

// recordCombatAttackLocked files one declared attack under the active
// player's current combat phase. Caller must hold g.mu.
func (g *Game) recordCombatAttackLocked(rec AttackRecord) {
	seat := g.Turn.ActiveSeat
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	p := g.Seats[seat]
	next := make([]AttackRecord, len(p.CombatAttacks), len(p.CombatAttacks)+1)
	copy(next, p.CombatAttacks)
	p.CombatAttacks = append(next, rec)
}

// PlayersAttackedDuringLastCombatForEffect is the players the object `c`
// attacked during its controller's last combat, each once (an attack on a
// planeswalker or a battle is not an attack on a player, CR 508.1b). It
// counts only if this very object attacked then (same instance, same
// epoch, CR 400.7), so a creature that was not on the battlefield, or was
// tapped, or did not attack then answers none.
//
// Caller must hold g.mu.
func (g *Game) PlayersAttackedDuringLastCombatForEffect(c *Card) []uuid.UUID {
	if c == nil {
		return nil
	}
	p := g.playerByIDLocked(c.Controller)
	if p == nil {
		return nil
	}
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, a := range p.LastCombatAttacks {
		if a.Attacker != c.InstanceID || a.Epoch != c.ObjectEpoch || seen[a.Defender] {
			continue
		}
		if g.playerByIDLocked(a.Defender) == nil {
			continue
		}
		seen[a.Defender] = true
		out = append(out, a.Defender)
	}
	return out
}
