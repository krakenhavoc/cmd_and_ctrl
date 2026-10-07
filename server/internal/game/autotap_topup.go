package game

import (
	"strings"

	"github.com/google/uuid"
)

// autotap_topup.go is ADR 0118 §1's pool top-up, which amends ADR 0011
// §7. The auto-tapper plans only what the floating pool is missing, so
// mana already in the pool is spent first and the plan pays the rest.
//
// Before this, every auto-tap payer asked one question of the pool,
// "does it cover the WHOLE cost?", and when the answer was no it asked
// the planner for the whole cost as if the pool were empty. With {G}
// floating and one untapped Forest, a {1}{G} creature was refused: the
// pool alone was short and the Forest alone was short, and nobody
// asked whether the two together paid. The legal-move enumerator asked
// the same two questions, so the bot and the ready ring called the
// card uncastable.
//
// One function now answers it for every payer, autoTapTopUpLocked:
//
//   - the cast (applyAutoTapLocked);
//   - an activation, a special action and an attack tax, which all pay
//     through payAbilityManaCostLocked;
//   - a pay-unless prompt (payCostLocked);
//   - the attack-tax affordability check (attackTaxAffordableLocked);
//   - and, through AutoTapTopUpForEffectExcluding, the enumerator's
//     canPayExcluding and its attack-tax probe, and the lobby's
//     /autotap preview through AutoTapPlanToppingUp.
//
// The planner itself is unchanged and still knows nothing about the
// pool. What changes is the cost it is asked to plan: the shortfall
// poolShortfalls computes. It stays read-only, and the plan still
// materialises under the payer's write lock. materializePlanLocked
// colour-picks against the shortfall rather than the whole cost, since
// the shortfall is what the plan was built to pay.
//
// What may pay from the pool is decided by the payment's own spend
// context (#352), so Ancient Ziggurat's creature-only {G} is credited
// toward a creature spell and toward nothing else, exactly as the
// payment that follows will spend it.

// maxPoolShortfalls caps how many ways of crediting the pool against
// a cost are offered to the planner. Only a hybrid or a widened symbol
// gives more than one way at all, so every ordinary cost has exactly
// one shortfall; the cap bounds a pathological cost of many hybrids,
// where the first shortfalls found are the ones that spend the most
// pool mana.
const maxPoolShortfalls = 16

// poolShortfalls returns what an auto-tap plan has to pay once the
// floating pool has paid what it can of `cost` under `ctx`, each as a
// ParsedCost the planner can be asked for: the colored symbols no
// spendable token was credited to, in the cost's own order, and the
// generic left over after every remaining token paid for it, with X
// already multiplied in (XSlots is zero; plan it with xValue 0).
//
// Usually there is exactly one answer. A single-colour symbol is
// credited first, from a token of its colour, because a token given to
// a single-colour symbol rather than to a hybrid that also takes it
// leaves the hybrid behind, which more sources can pay. A token is
// credited to a colored symbol before generic for the same reason:
// generic is the easiest thing to leave a plan. A hybrid symbol is
// where the pool can be spent more than one way ({G/U} or {G/W} out of
// one {G}), and which way is right depends on the board, so every way
// is returned, the ones that leave the plan less to pay first, and the
// caller takes the first the planner can fund. A shortfall that asks
// for at least as much as another, symbol for symbol, is dropped,
// since any plan that pays it pays the other.
//
// Returns nil when the pool covers the cost on its own; that is the
// caller's shortcut, and nothing needs planning.
func poolShortfalls(pool ManaPool, cost ParsedCost, xValue int, ctx ManaSpendContext) []ParsedCost {
	if pool.CanPayFor(cost, xValue, ctx) {
		return nil
	}
	counts := map[string]int{}
	// #2170: mana that can't pay generic costs is credited to a
	// coloured symbol before anything else (it can pay nothing else),
	// is never counted toward the generic demand in `total`, and is
	// counted apart in ngCounts / ngTotal so the credit can say which
	// kind it spent.
	ngCounts := map[string]int{}
	total, ngTotal := 0, 0
	for _, tok := range pool {
		if ctx.allows(tok.Restrictions) {
			counts[tok.Color]++
			if tok.noGeneric() {
				ngCounts[tok.Color]++
				ngTotal++
			} else {
				total++
			}
		}
	}
	// takeCredit / giveCredit move one token of colour c in and out of
	// the books, no-generic first.
	takeCredit := func(c string) (ng bool) {
		counts[c]--
		if ngCounts[c] > 0 {
			ngCounts[c]--
			ngTotal--
			return true
		}
		total--
		return false
	}
	giveCredit := func(c string, ng bool) {
		counts[c]++
		if ng {
			ngCounts[c]++
			ngTotal++
		} else {
			total++
		}
	}
	unpaid := make([]bool, len(cost.Required))
	var flexible []int
	for i, req := range cost.Required {
		if req.AnyMana || len(req.Options) != 1 {
			flexible = append(flexible, i)
			continue
		}
		if c := req.Options[0]; counts[c] > 0 {
			takeCredit(c)
			continue
		}
		unpaid[i] = true
	}
	need := cost.Generic + cost.XSlots*xValue

	var out []ParsedCost
	var credit func(k int)
	credit = func(k int) {
		if len(out) >= maxPoolShortfalls {
			return
		}
		if k == len(flexible) {
			out = append(out, shortfallCost(cost, unpaid, need-total))
			return
		}
		i := flexible[k]
		for _, c := range poolCreditColors(cost.Required[i]) {
			if counts[c] == 0 {
				continue
			}
			ng := takeCredit(c)
			credit(k + 1)
			giveCredit(c, ng)
		}
		unpaid[i] = true
		credit(k + 1)
		unpaid[i] = false
	}
	credit(0)
	return undominatedShortfalls(out)
}

// poolCreditColors is the colours of token that may be credited to one
// colored symbol, in the order they are tried: the colours it prints,
// and for a widened slot (AnyMana, #1589) every other mana after them,
// as the pool solver pays a widened slot with its printed colour first
// and with anything in the generic pass otherwise.
func poolCreditColors(req ColorRequirement) []string {
	all := []string{"W", "U", "B", "R", "G", "C"}
	if !req.AnyMana && len(req.Options) > 0 {
		return req.Options
	}
	out := append([]string(nil), req.Options...)
	for _, c := range all {
		if !containsFold(out, c) {
			out = append(out, c)
		}
	}
	return out
}

// shortfallCost is `cost` reduced to the colored symbols marked unpaid
// and `generic` generic mana (never below zero), with X folded into
// that generic. Every other field is kept, so the planner sees the
// same symbols the payment will.
func shortfallCost(cost ParsedCost, unpaid []bool, generic int) ParsedCost {
	out := cost
	out.Required = nil
	for i, req := range cost.Required {
		if unpaid[i] {
			out.Required = append(out.Required, req)
		}
	}
	if generic < 0 {
		generic = 0
	}
	out.Generic = generic
	out.XSlots = 0
	return out
}

// undominatedShortfalls drops every shortfall that asks for at least as
// much as another one (the same or more generic, and every colored
// symbol the other has), keeping the first of two equal ones, and
// orders what is left by how much it asks for, fewest symbols first.
// Stable, so the order credit() found them in breaks ties.
func undominatedShortfalls(in []ParsedCost) []ParsedCost {
	if len(in) <= 1 {
		return in
	}
	keep := make([]bool, len(in))
	for i := range in {
		keep[i] = true
		for j := range in {
			if i == j || !shortfallCovers(in[j], in[i]) {
				continue
			}
			// in[j] asks for no more than in[i]. Drop in[i] unless the
			// two are equal and in[i] came first.
			if !shortfallCovers(in[i], in[j]) || j < i {
				keep[i] = false
				break
			}
		}
	}
	var out []ParsedCost
	for i, c := range in {
		if keep[i] {
			out = append(out, c)
		}
	}
	size := func(c ParsedCost) int { return c.Generic + len(c.Required) }
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && size(out[j]) < size(out[j-1]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// shortfallCovers reports whether `small` asks for no more than `big`:
// no more generic, and each of its colored symbols matched by a
// distinct identical symbol of big's.
func shortfallCovers(small, big ParsedCost) bool {
	if small.Generic > big.Generic {
		return false
	}
	left := map[string]int{}
	for _, r := range big.Required {
		left[requirementKey(r)]++
	}
	for _, r := range small.Required {
		k := requirementKey(r)
		if left[k] == 0 {
			return false
		}
		left[k]--
	}
	return true
}

// requirementKey names one colored symbol for shortfallCovers: its
// printed shape, and whether a spend grant widened it.
func requirementKey(r ColorRequirement) string {
	var b strings.Builder
	b.WriteString(r.String())
	if r.AnyMana {
		b.WriteString("+any")
	}
	return b.String()
}

// autoTapTopUpLocked is the one auto-tap entry point every payer uses
// (see the file comment). It returns the plan and the cost the plan
// was built to pay, which is what materializePlanLocked colour-picks
// against. A pool that covers `cost` under `ctx` on its own returns
// an empty plan and true, without planning: auto-tap is idempotent on
// a funded pool.
//
// `cost` is the cost as the payer will pay it (after the spend-grant
// widening and the Phyrexian strike), and `xValue` its announced X.
// Read-only. Caller must hold g.mu.
func (g *Game) autoTapTopUpLocked(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	ctx ManaSpendContext,
	excluded map[uuid.UUID]bool,
	prefer ManaSourceKinds,
) (tapPlan, ParsedCost, bool) {
	p := g.playerByIDLocked(controller)
	if p == nil {
		plan, ok := g.autoTapPreferringLocked(controller, cost, xValue, excluded, prefer)
		return plan, cost, ok
	}
	if p.ManaPool.CanPayFor(cost, xValue, ctx) {
		return nil, ParsedCost{}, true
	}
	for _, short := range poolShortfalls(p.ManaPool, cost, xValue, ctx) {
		if plan, ok := g.autoTapPreferringLocked(controller, short, 0, excluded, prefer); ok {
			return plan, short, true
		}
	}
	return nil, ParsedCost{}, false
}

// AutoTapTopUpForEffectExcluding reports whether `controller` can pay
// `cost` from the floating pool and an auto-tap plan together (ADR
// 0118 §1), the question the engine's auto-tap payers ask. It is the
// legal-move enumerator's affordability probe, so the move list, the
// ready ring and the payment agree.
//
// `cost` is the cost as paid (CostAsPaidByForEffect), and `spend` the
// context the payment spends the pool under. For callers already
// under g.mu (the enumerator runs inside ReadSnapshot). Read-only.
func (g *Game) AutoTapTopUpForEffectExcluding(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	spend ManaSpendContext,
	excluded map[uuid.UUID]bool,
) bool {
	_, _, ok := g.autoTapTopUpLocked(controller, cost, xValue, spend, excluded, 0)
	return ok
}

// AutoTapPlanToppingUp is AutoTapPlanPreferringExcluding with the pool
// top-up: the sources the engine's auto-tapper would spend once the
// floating pool has paid what it can, each described. An empty plan
// with true means the pool covers the cost on its own. The lobby's
// /autotap preview is the caller, so the preview names exactly the
// taps the auto-tapped cast will make.
func (g *Game) AutoTapPlanToppingUp(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	spend ManaSpendContext,
	excluded map[uuid.UUID]bool,
	prefer ManaSourceKinds,
) ([]AutoTapPlanEntry, bool) {
	var (
		out []AutoTapPlanEntry
		ok  bool
	)
	g.ReadSnapshot(func() {
		var plan tapPlan
		plan, _, ok = g.autoTapTopUpLocked(controller, cost, xValue, spend, excluded, prefer)
		if ok {
			out = g.describePlanLocked(controller, plan)
		}
	})
	return out, ok
}

// planFundsLocked reports whether carrying out `plan` would leave
// `p`'s pool able to pay `cost` under `ctx` (#2461). It carries the
// plan out on a throwaway clone of the game (cloneLocked, the undo
// stack's deep copy) and asks the clone's pool, so the live game is
// untouched whatever the answer.
//
// Every auto-tap payer asks it between planning and materialising, so
// a payment the plan cannot fund is refused BEFORE anything is tapped:
// CastSpellParams.AutoTap's all-or-nothing promise. The planner and
// the executor are two halves that have to agree on which slot pays
// which pip, and when they did not (#2461: a bounce land's {U} booked
// twice) the cast tapped the land, floated its mana and then refused.
// The planner is fixed; this keeps the next disagreement from
// stranding a board.
//
// The clone runs the real executor, so it sees what the live run will
// see: the CR 614 window on the production, the triggered mana
// abilities, a sacrifice, a pain rider. An effect error raised inside
// it is logged and counted for the clone as well; nothing else leaves
// the clone.
//
// An empty plan is not cloned for. Caller must hold g.mu.
func (g *Game) planFundsLocked(p *Player, plan tapPlan, short, cost ParsedCost, xValue int, ctx ManaSpendContext) bool {
	if len(plan) == 0 {
		return p.ManaPool.CanPayFor(cost, xValue, ctx)
	}
	trial := g.cloneLocked()
	tp := trial.playerByIDLocked(p.ID)
	if tp == nil {
		return false
	}
	trial.materializePlanLocked(tp, plan, short)
	return tp.ManaPool.CanPayFor(cost, xValue, ctx)
}
