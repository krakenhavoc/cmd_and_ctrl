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

// blocksParams is the set-shaped declare_blockers payload (#750). A
// block that only exists as a group — the two creatures a menace
// attacker takes — cannot be sent as two declare_blocker actions,
// because the first would be refused for too_few_blockers.
type blocksParams struct {
	Blocks []blockParams `json:"blocks"`
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
		// #750, ADR 0045 addendum Decision 14: the options come from
		// the engine's own generator, which runs the same validator
		// DeclareBlockers runs. That is what keeps the enumerator, the
		// #328 auto-pass signal and the verb in agreement (ADR 0045
		// §3) now that block legality includes a COUNT: a lone block
		// on a menace attacker is refused, so it is never offered,
		// and the two-creature block that IS legal is offered as one
		// grouped move rather than as two singles the engine would
		// reject one at a time.
		//
		// BlockerEligible, the per-card test the wire's
		// block_decision_seats signal uses, is applied inside the
		// generator for the same agreement reason.
		for _, opt := range g.BlockOptionsLocked(e.seat, e.opts.MaxExpansionPerSource) {
			if len(opt.Blocks) == 0 {
				continue
			}
			first := opt.Blocks[0]
			atk := cardByID(g, first.Attacker)
			if atk == nil {
				continue
			}
			if len(opt.Blocks) == 1 {
				blk := cardByID(g, first.Blocker)
				if blk == nil {
					continue
				}
				e.add(Move{
					Type:   TypeDeclareBlocker,
					Player: e.seat,
					Kind:   KindBlock,
					Label:  "Block " + atk.Name + " with " + blk.Name,
					Source: blk.InstanceID,
					Params: mustJSON(blockParams{Blocker: blk.InstanceID.String(), Attacker: atk.InstanceID.String()}),
				})
				continue
			}
			set := blocksParams{Blocks: make([]blockParams, 0, len(opt.Blocks))}
			names := make([]string, 0, len(opt.Blocks))
			ok := true
			for _, d := range opt.Blocks {
				blk := cardByID(g, d.Blocker)
				if blk == nil {
					ok = false
					break
				}
				set.Blocks = append(set.Blocks, blockParams{
					Blocker:  d.Blocker.String(),
					Attacker: d.Attacker.String(),
				})
				names = append(names, blk.Name)
			}
			if !ok {
				continue
			}
			e.add(Move{
				Type:   TypeDeclareBlockers,
				Player: e.seat,
				Kind:   KindBlock,
				Label:  "Block " + atk.Name + " with " + joinNames(names),
				Source: opt.Blocks[0].Blocker,
				Params: mustJSON(set),
			})
		}
	}
}

// cardByID finds a battlefield permanent by instance ID. Caller holds
// the enumerator's read lock.
func cardByID(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// joinNames renders a group block's creatures for the move label:
// "A and B", "A, B and C".
func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	}
	out := ""
	for i, n := range names[:len(names)-1] {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out + " and " + names[len(names)-1]
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
