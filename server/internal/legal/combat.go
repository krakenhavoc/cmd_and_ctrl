package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

type attackParams struct {
	Attacker string `json:"attacker"`
	Target   string `json:"target"`

	// AutoTap lets the declaration tap lands for the CR 508.1a attack
	// tax (ADR 0080). Set only when the seat's pool alone cannot cover
	// the tax but the tapper can, exactly as the cast arm sets it —
	// the Move doc's promise is that Params is EXACTLY the payload
	// that performs the move, so a move offered on the strength of the
	// tapper has to carry permission to use it.
	AutoTap bool `json:"auto_tap,omitempty"`
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
		// #1571 / CR 508.1d: while the declaration owes a requirement,
		// the attacks that would answer it. Nil at every table with no
		// requirement to meet, which is nearly all of them.
		owed := g.MustAttackForEffect()
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
				decl := []game.AttackDeclaration{{Attacker: c.InstanceID, Target: t.ID}}
				// #1507, CR 508.1c: a count limit (Silent Arbiter,
				// Crawlspace) the declaration verb would refuse is
				// not offered — the same function, so the two cannot
				// disagree (#544). Moves are one creature at a time,
				// so a seat under Silent Arbiter is offered attacks
				// until one creature is attacking, then none; under
				// Crawlspace, attacks at that player until two are.
				if g.AttackLimitRefusalForEffect(decl) != nil {
					continue
				}
				// #1571, CR 508.1d: a declaration that makes a
				// requirement unobeyable (a goaded creature at its
				// goader while another opponent is open; a creature
				// with no requirement into the one slot Zurgo needs)
				// is refused by the verb, so it is not offered — the
				// same function (#544).
				if g.AttackRequirementRefusalForEffect(decl) != nil {
					continue
				}
				// ADR 0080 / #1063: the CR 508.1a attack tax. Priced
				// through the SAME function the engine charges
				// (PriceAttackDeclarationForEffect), so the price the
				// bot is offered and the price it is charged cannot
				// disagree, and dropped outright when the seat cannot
				// pay it — the #544 rule, that a bot is never offered
				// a move the engine refuses.
				//
				// Priced per (attacker, target) pair because that is
				// the granularity the move declares: the engine
				// charges one declaration verb call at a time, so
				// three separate attacks under Propaganda pay {2}
				// three times, and each re-enumeration prices the next
				// one against the mana the last one left.
				price := g.PriceAttackDeclarationForEffect(decl)
				autoTap := false
				if !price.IsFree() {
					if !e.p.ManaPool.CanPayFor(price.Total, 0, game.ManaSpendContext{}) {
						// Only the tapper can cover it, so the move
						// has to say so — and if the tapper cannot
						// either, the move is not offered at all.
						if _, ok := e.g.AutoTapForCostForEffectExcluding(e.seat, price.Total, 0, nil); !ok {
							continue
						}
						autoTap = true
					}
				}
				e.add(Move{
					Type:   TypeDeclareAttacker,
					Player: e.seat,
					Kind:   KindAttack,
					Label:  attackMoveLabel(g, t, c.Name, price),
					Source: c.InstanceID,
					// #1571: an attack that answers an owed
					// requirement is the declaration's unconditional
					// answer while the pass is withheld — nothing else
					// can act in the active player's priority window to
					// make the engine refuse it — so a seat whose
					// policy declines is not left holding the table.
					AlwaysLegal: owedPair(owed, c.InstanceID, t.ID),
					Cost:        withAttackTax(nil, price.Cost),
					Params: mustJSON(attackParams{
						Attacker: c.InstanceID.String(),
						Target:   t.ID.String(),
						AutoTap:  autoTap,
					}),
				})
			}
		}
	case game.StepDeclareBlockers:
		// Attackers pointed at this seat.
		// S27: "attacking this seat" is the DEFENDING player of the
		// declaration, not a bare id match — an attack on a
		// planeswalker names the walker and is defended by its
		// controller. #1364: the ATTACKER-based read, so a creature
		// whose planeswalker or battle has left combat is still
		// offered to the player who was defending it (CR 506.4c) —
		// the same answer the generator below and the verb give.
		var attackers []*game.Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.AttackingTarget == uuid.Nil {
				continue
			}
			if g.DefendingPlayerForAttackerForEffect(c.InstanceID) == e.seat {
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

// attackRequirementOwed reports whether `seat` is the active player in
// declare_attackers with a CR 508.1d requirement their declaration
// could still obey and does not (#1571) — the case the engine refuses
// their pass in. Caller holds the enumerator's read lock.
func attackRequirementOwed(g *game.Game, seat uuid.UUID) bool {
	if g.Turn.Step != game.StepDeclareAttackers || !isActiveSeat(g, seat) {
		return false
	}
	return g.AttackRequirementsUnmetForEffect() != nil
}

// owedPair reports whether (attacker, target) is one of the attacks
// that answers an owed requirement.
func owedPair(owed map[uuid.UUID][]uuid.UUID, attacker, target uuid.UUID) bool {
	for _, t := range owed[attacker] {
		if t == target {
			return true
		}
	}
	return false
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
// attackMoveLabel is the move's human line, with the CR 508.1a tax
// named when there is one. A bot's decision log and the client's move
// list both read it, and "Attack Alice with Bear" reads as free
// whether it is or not.
func attackMoveLabel(g *game.Game, t game.AttackTargetRef, attacker string, price game.AttackTaxPrice) string {
	label := "Attack " + attackTargetLabel(g, t) + " with " + attacker
	if price.IsFree() {
		return label
	}
	return label + " (pays " + price.Cost + ")"
}

// withAttackTax adds the CR 508.1a mana price to a (possibly nil)
// MoveCost, returning a fresh value so no two moves share one — the
// contract withCounterPrice has, for the same reason.
func withAttackTax(c *MoveCost, mana string) *MoveCost {
	if mana == "" {
		return c
	}
	out := MoveCost{}
	if c != nil {
		out = *c
		out.Counters = append([]CounterPrice(nil), c.Counters...)
	}
	out.Mana = mana
	return &out
}

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
