package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declaration_control.go — moves under a "you choose which creatures
// attack / block" effect (#4154, Master Warcraft; game's
// declaration_control.go).
//
//   - The declarer is offered the declaration it is making for another
//     seat, enumerated as that seat, each move's Player set to it
//     (controlledDeclarationMoves): the attack moves and the finishing
//     pass for the active player's declaration, the block moves and
//     finish_blocks for a defender's.
//   - The active player, in their half of a chosen attack declaration,
//     is offered a re-point of each chosen attacker, the confirmation
//     that pays the costs, and (in run) the pass when the engine takes
//     it (attackTargetsMoves).

// declarationAttackersPayload is declare_attackers' payload, as the
// actions layer reads it.
type declarationAttackersPayload struct {
	Attackers     []attackParams `json:"attackers"`
	AutoTap       bool           `json:"auto_tap,omitempty"`
	PhyrexianLife int            `json:"phyrexian_life,omitempty"`
}

// chosenAttackTaxCost is a chosen attack's CR 508.1a price as a move's
// cost, nil when free.
func chosenAttackTaxCost(mana string) *MoveCost {
	if mana == "" {
		return nil
	}
	return &MoveCost{Mana: mana}
}

// controlledDeclarationMoves appends the moves of each declaration e's
// seat is making for another seat right now. Caller must hold g.mu.
func (e *enumerator) controlledDeclarationMoves() {
	for _, p := range e.g.Seats {
		if p == nil || p.ID == e.seat || p.Eliminated {
			continue
		}
		if e.g.ControlStageOfPlayerLocked(p.ID) != game.ControlDeclaring ||
			e.g.ControllerOfPlayerLocked(p.ID) != e.seat {
			continue
		}
		sub := enumeratorAs(e.g, p.ID, e.opts)
		if sub == nil {
			continue
		}
		sub.report = e.report
		if e.g.Turn.Step == game.StepDeclareAttackers && holdsPriority(e.g, p.ID) && !attackRequirementOwed(e.g, p.ID) {
			sub.add(Move{
				Type:        TypePassPriority,
				Player:      p.ID,
				Kind:        KindPass,
				Label:       "Done choosing attackers",
				AlwaysLegal: true,
			})
		}
		sub.combatMoves()
		for i := range sub.out {
			sub.out[i].Player = p.ID
		}
		e.out = append(e.out, sub.out...)
		e.cuts = append(e.cuts, sub.cuts...)
		e.added += sub.added
	}
}

// attackTargetsMoves is the active player's half of a chosen attack
// declaration: each chosen attacker pointed at each other target it may
// attack (the engine re-points it and charges nothing yet), and, when
// the targets as they stand cost something the seat can pay, the
// confirmation that pays it all. The pass, which accepts the targets
// and declines any cost, is run's (AttackTargetsPassRefusalForEffect).
// Caller must hold g.mu.
func (e *enumerator) attackTargetsMoves() {
	g := e.g
	staged := g.StagedAttackersForEffect()
	if len(staged) == 0 {
		return
	}
	for _, c := range staged {
		if !e.enter(scope{source: c.InstanceID}) {
			continue
		}
		for _, t := range g.AttackTargetsForAttackerForEffect(c) {
			if t.ID == c.AttackingTarget {
				continue
			}
			decl := []game.AttackDeclaration{{Attacker: c.InstanceID, Target: t.ID}}
			if g.AttackLimitRefusalForEffect(decl) != nil || g.AttackRequirementRefusalForEffect(decl) != nil {
				continue
			}
			// The price is the one the confirmation will ask for this
			// target; the re-point itself charges nothing.
			price := g.PriceAttackDeclarationForEffect(decl)
			e.add(Move{
				Type:   TypeDeclareAttacker,
				Player: e.seat,
				Kind:   KindAttack,
				Label:  attackMoveLabel(g, t, c.Name, price),
				Source: c.InstanceID,
				Cost:   chosenAttackTaxCost(price.Cost),
				Params: mustJSON(attackParams{
					Attacker: c.InstanceID.String(),
					Target:   t.ID.String(),
				}),
			})
		}
	}
	e.leave()
	decls := make([]game.AttackDeclaration, 0, len(staged))
	payload := declarationAttackersPayload{Attackers: make([]attackParams, 0, len(staged))}
	for _, c := range staged {
		decls = append(decls, game.AttackDeclaration{Attacker: c.InstanceID, Target: c.AttackingTarget})
		payload.Attackers = append(payload.Attackers, attackParams{
			Attacker: c.InstanceID.String(),
			Target:   c.AttackingTarget.String(),
		})
	}
	price := g.PriceAttackDeclarationForEffect(decls)
	if price.IsFree() {
		return
	}
	total := g.CostAsPaidByForEffect(e.seat, game.ManaSpendContext{}, price.Total, 0)
	autoTap, phyLife, ok := e.attackTaxPayment(total)
	if !ok {
		return
	}
	payload.AutoTap, payload.PhyrexianLife = autoTap, phyLife
	e.add(Move{
		Type:   TypeDeclareAttackers,
		Player: e.seat,
		Kind:   KindAttack,
		Label:  "Confirm the attacks (pays " + price.Cost + ")" + attackLifeLabel(phyLife),
		Source: uuid.Nil,
		Cost:   withPhyrexianLife(chosenAttackTaxCost(price.Cost), phyLife),
		Params: mustJSON(payload),
		// The pass is refused while a cost could be avoided by a free
		// re-point; this is then an answer nothing else can refuse.
		AlwaysLegal: g.AttackTargetsPassRefusalForEffect() != nil,
	})
}
