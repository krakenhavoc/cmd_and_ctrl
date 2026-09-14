package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

type attackParams struct {
	Attacker string `json:"attacker"`
	Target   string `json:"target"`
}

type blockParams struct {
	Blocker  string `json:"blocker"`
	Attacker string `json:"attacker"`
}

// combatMoves enumerates per-creature attack and block declarations.
// Neither is priority-gated in the engine — the step and the card's
// controller are the whole check — so a defending seat can block
// while the active player still holds priority.
//
// Attack and block sets are combinatorial; a policy composes a full
// declaration from these per-creature moves, re-enumerating after
// each one (a declared attacker drops out of the list because the
// engine taps it; a declared blocker drops out because it is already
// blocking).
//
// Where the engine is lax, this is strict (CR 508.1a / 509.1a): an
// attacker must be untapped and controlled by the active seat, and
// may only attack an opponent who is still in the game; a blocker
// must be untapped, controlled by the defending seat, and may only
// block an attacker that is attacking that seat.
func (e *enumerator) combatMoves() {
	g := e.g
	switch g.Turn.Step {
	case game.StepDeclareAttackers:
		if !isActiveSeat(g, e.seat) {
			return
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			// S24: the attacker-side eligibility rule is
			// game.AttackerEligible and the ENGINE's bulk
			// DeclareAttackers runs the same function. This list used
			// to be spelled out here — controller, creature,
			// untapped, not already declared, no defender, not
			// summoning sick — and spelled out again in the engine,
			// which is the arrangement #544 hung a table with: the
			// enumerator offered a move the engine refused, and a
			// seat that owes a decision is enumerated that decision's
			// answers and nothing else, so there was nothing else to
			// do. "Can't attack" (Pacifism) joins the list inside
			// that one function and both sides get it at once.
			if !game.AttackerEligible(c, e.seat) {
				continue
			}
			// S27: an attacker may be declared against a player, a
			// planeswalker or a battle (CR 508.1d), so the target set
			// comes from the engine rather than from a walk over the
			// seats. Enumerating only players would have left the bot
			// unable to see a lethal swing at a planeswalker.
			for _, t := range g.AttackTargetsForEffect(e.seat) {
				e.add(Move{
					Type:   TypeDeclareAttacker,
					Player: e.seat,
					Kind:   KindAttack,
					Label:  "Attack " + attackTargetLabel(g, t) + " with " + c.Name,
					Source: c.InstanceID,
					Params: mustJSON(attackParams{Attacker: c.InstanceID.String(), Target: t.ID.String()}),
				})
			}
		}
	case game.StepDeclareBlockers:
		// Attackers pointed at this seat.
		// S27: "attacking this seat" is the DEFENDING player of the
		// declaration, not a bare id match — an attack on a
		// planeswalker names the walker and is defended by its
		// controller.
		var attackers []*game.Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.AttackingTarget == uuid.Nil {
				continue
			}
			if g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == e.seat {
				attackers = append(attackers, c)
			}
		}
		if len(attackers) == 0 {
			return
		}
		for i := range g.Battlefield.Cards {
			b := &g.Battlefield.Cards[i]
			// #328: the same per-card eligibility test the wire's
			// block_decision_seats signal uses, so the enumerator and
			// the auto-pass guard can never disagree about whether a
			// seat has a block available.
			if !game.BlockerEligible(b, e.seat) {
				continue
			}
			for _, a := range attackers {
				if !game.CanBlock(a, b) {
					continue
				}
				e.add(Move{
					Type:   TypeDeclareBlocker,
					Player: e.seat,
					Kind:   KindBlock,
					Label:  "Block " + a.Name + " with " + b.Name,
					Source: b.InstanceID,
					Params: mustJSON(blockParams{Blocker: b.InstanceID.String(), Attacker: a.InstanceID.String()}),
				})
			}
		}
	}
}

// attackTargetLabel renders an attack target for the move's human
// label: a seat's name, or a permanent's card name.
func attackTargetLabel(g *game.Game, t game.AttackTargetRef) string {
	if t.Kind == game.AttackTargetPlayer {
		for _, p := range g.Seats {
			if p != nil && p.ID == t.ID {
				return p.Name
			}
		}
		return "a player"
	}
	if c, ok := g.LookupCardForEffect(t.ID); ok {
		return c.Name
	}
	return "a permanent"
}
