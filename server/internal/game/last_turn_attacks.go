package game

import "github.com/google/uuid"

// last_turn_attacks.go — what a player's creatures did during that
// player's LAST turn (ADR 0108 §6, #1882).
//
// TurnTally.Attacks is the current turn's record and is replaced as each
// turn begins, so nothing in it survives into the turn that asks "did it
// attack during your last turn?" (Goblin Rock Sled, Giant Turtle, Tangle
// Kelp, Avenge, O-Kagachi). As a turn ends, the attacks it recorded are
// copied onto the player whose turn it was, replacing that player's
// previous record: Player.LastTurnAttacks.
//
// "Your last turn" is the most recent turn YOU took, so:
//   - an extra turn is a turn (CR 500.7) and replaces the record, which
//     means the turn that follows reads the extra turn, as printed;
//   - a turn that is not taken writes nothing: a seat the rotation steps
//     over (CR 800.4k) never begins one, and the record is left as it was;
//   - an empty record is written too. A turn in which you attacked with
//     nothing is your last turn, and the turn before it is forgotten.
//
// Only declarations are recorded (CR 508.4: a creature put onto the
// battlefield attacking was never declared as an attacker), and each is
// keyed by object (CR 400.7), so a creature that left the battlefield and
// returned has no history.

// recordLastTurnAttacksLocked copies the ending turn's attacks onto the
// player whose turn it was. It runs as the turn ends, before the rotation
// replaces g.Turn and the tally. Caller must hold g.mu.
func (g *Game) recordLastTurnAttacksLocked() {
	seat := g.Turn.ActiveSeat
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	// A fresh slice, never a shared backing array: the undo clone holds
	// the old record (the reason turn_tally.go gives for Attacks).
	g.Seats[seat].LastTurnAttacks = append([]AttackRecord(nil), g.TurnTally.Attacks...)
}

// AttackedDuringControllersLastTurn reports whether the object `c` was
// declared as an attacker during its controller's last turn: "if it
// attacked during your last turn" (Goblin Rock Sled, Giant Turtle) and
// "if it attacked during its controller's last turn" (Tangle Kelp). The
// controller is the one now, so a creature that changed hands is asked
// of its new controller's last turn; it counts only if this very object
// attacked then (same instance, same epoch, CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) AttackedDuringControllersLastTurn(c *Card) bool {
	if c == nil {
		return false
	}
	p := g.playerByIDLocked(c.Controller)
	if p == nil {
		return false
	}
	for _, a := range p.LastTurnAttacks {
		if a.Attacker == c.InstanceID && a.Epoch == c.ObjectEpoch {
			return true
		}
	}
	return false
}

// AttackedYouDuringTheirLastTurn reports whether `player` attacked `you`
// during player's last turn: "if that player attacked you during their
// last turn" (O-Kagachi). An attack on a planeswalker or a battle you
// control is not an attack on you (CR 508.1b names the player,
// planeswalker or battle each attacker attacks).
//
// Caller must hold g.mu.
func (g *Game) AttackedYouDuringTheirLastTurn(player, you uuid.UUID) bool {
	p := g.playerByIDLocked(player)
	if p == nil || you == uuid.Nil {
		return false
	}
	for _, a := range p.LastTurnAttacks {
		if a.Defender == you {
			return true
		}
	}
	return false
}

// AnyPlayerAttackedYouDuringTheirLastTurn is the "a player" form: some
// other player attacked you during their last turn (Avenge).
//
// Caller must hold g.mu.
func (g *Game) AnyPlayerAttackedYouDuringTheirLastTurn(you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && p.ID != you && g.AttackedYouDuringTheirLastTurn(p.ID, you) {
			return true
		}
	}
	return false
}
