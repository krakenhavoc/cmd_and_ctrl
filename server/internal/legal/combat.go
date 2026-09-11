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
			if c.Controller != e.seat || !c.IsCreature() || c.Tapped {
				continue
			}
			if c.AttackingTarget != uuid.Nil {
				continue
			}
			if game.HasKeyword(c, "defender") || game.HasSummoningSickness(c) {
				continue
			}
			for _, opp := range g.Seats {
				if opp == nil || opp.ID == e.seat || opp.Eliminated {
					continue
				}
				e.add(Move{
					Type:   TypeDeclareAttacker,
					Player: e.seat,
					Kind:   KindAttack,
					Label:  "Attack " + opp.Name + " with " + c.Name,
					Source: c.InstanceID,
					Params: mustJSON(attackParams{Attacker: c.InstanceID.String(), Target: opp.ID.String()}),
				})
			}
		}
	case game.StepDeclareBlockers:
		// Attackers pointed at this seat.
		var attackers []*game.Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.AttackingTarget == e.seat {
				attackers = append(attackers, c)
			}
		}
		if len(attackers) == 0 {
			return
		}
		for i := range g.Battlefield.Cards {
			b := &g.Battlefield.Cards[i]
			if b.Controller != e.seat || !b.IsCreature() || b.Tapped {
				continue
			}
			if b.BlockingTarget != uuid.Nil {
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
